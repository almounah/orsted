package listeners

import (
	"context"
	"net/http"
	"orsted/server/utils"
	"time"
)

type LISTENER_HTTPS struct {
	Id       string
	Ip       string
	Port     string
	Type     string
	CertPath string
	KeyPath  string
	Srv      *http.Server
}

func NewHttpsListener(id string, ip string, port string, certPath string, keyPath string) *LISTENER_HTTPS {
	s := LISTENER_HTTPS{
		Id:       id,
		Ip:       ip,
		Port:     port,
		Type:     "https",
		CertPath: certPath,
		KeyPath:  keyPath,
	}
	return &s
}

func (s *LISTENER_HTTPS) StopListener() error {
    utils.PrintInfo("Stopping HTTPS Listener on server", s.Ip, s.Port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
    return s.Srv.Shutdown(ctx)
}

func (s *LISTENER_HTTPS) StartListener() error {
    utils.PrintInfo("Starting HTTPS Listener on server", s.Ip, s.Port)
	mux := http.NewServeMux()
    // Https handler are the same as HTTP handler
	addHttpHandler(mux)

	server := &http.Server{
		Addr:    s.Ip + ":" + s.Port,
		Handler: mux,
	}
    s.Srv = server
    err := s.Srv.ListenAndServeTLS(s.CertPath, s.KeyPath)
    if err != nil {
        utils.PrintInfo("Error in Serving HTTPS Listener ", err.Error())
    }
	return err
}

func (s *LISTENER_HTTP) GetListenerId() string {
    return s.Id
}
