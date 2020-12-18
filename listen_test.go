package httpserver_test

import (
	"github.com/normegil/closer"
	"github.com/normegil/connectionutils"
	"github.com/normegil/httpserver"
	"github.com/normegil/interval"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestServer_Listen_Errors(t *testing.T) {
	tests := []struct {
		Name string
		Port int
	}{
		{
			Name: "Negative port",
			Port: -1,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			server := httpserver.Server{
				Address: "0.0.0.0",
				Port:    test.Port,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("testerror"))
				}),
				Closer: closer.CloseErrorHandlerFunc(func(closer io.Closer) {
					t.Error(closer)
				}),
			}
			ctrl := server.Listen()
			defer shutdown(t, ctrl)

			checkServerErrors(t, ctrl, true)
		})
	}
}

func TestServer_Listen(t *testing.T) {
	tests := []struct {
		Name string
	}{
		{
			Name: "MyTest",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			address := "0.0.0.0"
			addr := connectionutils.SelectPort(net.ParseIP(address), *interval.MustParseIntervalInteger("[24000;24010]"))
			server := httpserver.Server{
				Address: addr.IP.String(),
				Port:    addr.Port,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("test"))
				}),
				Closer: closer.CloseErrorHandlerFunc(func(closer io.Closer) {
					if err := closer.Close(); nil != err {
						t.Error(err)
					}
				}),
			}
			ctrl := server.Listen()
			defer shutdown(t, ctrl)

			checkServerErrors(t, ctrl, false)

			client := &http.Client{}
			defer client.CloseIdleConnections()
			resp, err := client.Get("http://localhost:" + strconv.Itoa(addr.Port))
			if err != nil {
				t.Fatal(err)
			}

			if http.StatusOK != resp.StatusCode {
				t.Fatalf("Unexpected status: %s", resp.Status)
			}
		})
	}
}

func checkServerErrors(t testing.TB, ctrl *httpserver.ServerControl, expectedError bool) {
	for ctrl.Status() < httpserver.Listening {
		time.Sleep(5 * time.Millisecond)
	}

	select {
	case err := <-ctrl.Errors:
		status := ctrl.Status()
		if status != httpserver.Error {
			t.Errorf("status should be 'error' if an error occured. got: %d", status)
		}
		if !expectedError {
			t.Fatalf("unexpected error: %s", err.Error())
		}
	default:
		if expectedError {
			t.Errorf("error expected (%t) is not expected result (%t)", expectedError, false)
		}
	}
}

func shutdown(t testing.TB, controls *httpserver.ServerControl) {
	if err := controls.Shutdown(1000 * time.Millisecond); nil != err {
		t.Errorf("stopping server: %s", err)
	}
}
