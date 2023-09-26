package internal

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/snappy"
	"github.com/google/uuid"
	"github.com/net-byte/water"
	"github.com/xorgal/xtun-core/pkg/cache"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/counter"
	"github.com/xorgal/xtun-core/pkg/tun"
)

type RegisterDeviceRequest struct {
	DeviceId string `json:"id"`
}

type RegisterDeviceResponse struct {
	DeviceId string `json:"deviceId"`
	Server   string `json:"server"`
	Client   string `json:"client"`
}

type ServerConfigurationResponse struct {
	BufferSize int  `json:"bufferSize"`
	MTU        int  `json:"mtu"`
	Compress   bool `json:"compress"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
	Disconnecting
)

var (
	connectionState ConnectionState
	connMutex       sync.Mutex
	suspended       bool
)

func GetServerConfiguration(config config.Config) (ServerConfigurationResponse, error) {
	res, err := post(config, "/config", nil)
	if err != nil {
		return ServerConfigurationResponse{}, err
	}
	var result ServerConfigurationResponse
	err = json.Unmarshal(res, &result)
	if err != nil {
		return ServerConfigurationResponse{}, err
	}
	return result, nil
}

func GetIP(config config.Config) (RegisterDeviceRequest, RegisterDeviceResponse, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return RegisterDeviceRequest{}, RegisterDeviceResponse{}, err
	}
	config.DeviceId = id.String()
	payload := RegisterDeviceRequest{
		DeviceId: id.String(),
	}
	req, err := json.Marshal(payload)
	if err != nil {
		return payload, RegisterDeviceResponse{}, err
	}
	res, err := post(config, "/allocator/register", req)
	if err != nil {
		return payload, RegisterDeviceResponse{}, err
	}
	var result RegisterDeviceResponse
	err = json.Unmarshal(res, &result)
	if err != nil {
		return payload, RegisterDeviceResponse{}, err
	}
	return payload, result, nil
}

func StartClient(config config.Config, errCh chan<- error) {
	log.Println("Starting ws client...")
	setConnectionState(Connecting)
	suspended = false
	iface, err := tun.CreateTunInterface(config)
	if err != nil {
		log.Println(err)
		errCh <- err
		return
	}
	cache.GetCache().Set("iface", iface, 24*time.Hour)
	go tunToWs(config, iface)
	for {
		if suspended {
			return
		}
		conn, err := connect(config)
		if err != nil {
			log.Println(err)
			errCh <- err
		}
		if conn == nil {
			setConnectionState(Disconnected)
			time.Sleep(3 * time.Second)
			continue
		}
		setConnectionState(Connected)
		cache.GetCache().Set("wsconn", conn, 24*time.Hour)
		go wsToTun(config, conn, iface)
		ping(conn, config)
		cache.GetCache().Delete("wsconn")
		setConnectionState(Disconnected)
	}
}

func StopClient(config config.Config) error {
	log.Println("Stopping ws client...")
	setConnectionState(Disconnecting)
	if v, ok := cache.GetCache().Get("wsconn"); ok {
		conn := v.(net.Conn)
		if conn != nil {
			err := conn.Close()
			if err != nil {
				return err
			}
		}
	}
	if v, ok := cache.GetCache().Get("iface"); ok {
		iface := v.(*water.Interface)
		if iface != nil {
			err := iface.Close()
			if err != nil {
				return err
			}
		}
	}
	cache.GetCache().Delete("wsconn")
	cache.GetCache().Delete("iface")
	tun.ResetRoute(config)
	suspended = true
	setConnectionState(Disconnected)
	return nil
}

func connect(config config.Config) (net.Conn, error) {
	scheme := "ws"
	host := config.ServerAddr
	if config.Protocol == "wss" {
		scheme = "wss"
	}
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   "/ws",
	}
	header := make(http.Header)
	header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
	if config.Key != "" {
		header.Set("key", config.Key)
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}
	dialer := ws.Dialer{
		Header:    ws.HandshakeHeaderHTTP(header),
		Timeout:   time.Duration(120) * time.Second,
		TLSConfig: tlsConfig,
		NetDial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, config.ServerAddr)
		},
	}
	c, _, _, err := dialer.Dial(context.Background(), u.String())
	if err != nil {
		return nil, err
	}
	return c, err
}

func ping(wsconn net.Conn, config config.Config) {
	defer wsconn.Close()
	for {
		err := wsutil.WriteClientMessage(wsconn, ws.OpText, []byte("ping"))
		if err != nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
}

// wsToTun sends packets from ws to tun
func wsToTun(config config.Config, wsconn net.Conn, iface *water.Interface) {
	defer wsconn.Close()
	for {
		packet, err := wsutil.ReadServerBinary(wsconn)
		if err != nil {
			log.Print(err)
			break
		}
		if config.Compress {
			packet, _ = snappy.Decode(nil, packet)
		}
		_, err = iface.Write(packet)
		if err != nil {
			log.Print(err)
			break
		}
		counter.IncrReadBytes(len(packet))
	}
}

// tunToWs sends packets from tun to ws
func tunToWs(config config.Config, iface *water.Interface) {
	packet := make([]byte, config.BufferSize)
	for {
		n, err := iface.Read(packet)
		if err != nil {
			log.Print(err)
			break
		}
		if v, ok := cache.GetCache().Get("wsconn"); ok {
			b := packet[:n]
			if config.Compress {
				b = snappy.Encode(nil, b)
			}
			wsconn := v.(net.Conn)
			if err = wsutil.WriteClientBinary(wsconn, b); err != nil {
				log.Print(err)
				continue
			}
			counter.IncrWrittenBytes(n)
		}
	}
}

func setConnectionState(state ConnectionState) {
	connMutex.Lock()
	connectionState = state
	connMutex.Unlock()
}

// use this function to safely read isConnected
func GetConnectionState() ConnectionState {
	connMutex.Lock()
	defer connMutex.Unlock()
	return connectionState
}

func post(config config.Config, route string, body []byte) ([]byte, error) {
	client := getHttpClient(config)
	req, err := http.NewRequest("POST", "https://"+config.ServerAddr+route, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.Key != "" {
		req.Header.Set("key", config.Key)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return body, err
	} else {
		var errorResponse ErrorResponse
		err := json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, err
		} else {
			return nil, errors.New(errorResponse.Message)
		}
	}
}

func getHttpClient(config config.Config) http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
		},
		Timeout: time.Duration(120) * time.Second,
	}
	return *client
}
