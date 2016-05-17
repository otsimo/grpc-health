package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	healthClient       healthpb.HealthClient
	healthConn         *grpc.ClientConn
	remoteUrl          string        = ""
	serviceName        string        = ""
	secureConnection   bool          = true
	insecureSkipVerify bool          = false
	timeoutDur         time.Duration = time.Second
	Version            string        = "DEV"
)

func connectToRemote() {
	var opts []grpc.DialOption
	if secureConnection {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: "", InsecureSkipVerify: insecureSkipVerify})))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	c, err := grpc.Dial(remoteUrl, opts...)
	if err != nil {
		glog.Errorf("failed to dial grpc server, %v", err)
		return
	}
	healthConn = c
	healthClient = healthpb.NewHealthClient(healthConn)
}

// SetFlagsFromEnv parses all registered flags in the given flagset,
// and if they are not already set it attempts to set their values from
// environment variables. Environment variables take the name of the flag but
// are UPPERCASE, and any dashes are replaced by underscores. Environment
// variables additionally are prefixed by the given string followed by
// and underscore. For example, if prefix=PREFIX: some-flag => PREFIX_SOME_FLAG
func SetFlagsFromEnv(fs *flag.FlagSet, prefix string) error {
	var err error
	alreadySet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		alreadySet[f.Name] = true
	})
	fs.VisitAll(func(f *flag.Flag) {
		if !alreadySet[f.Name] {
			key := prefix + "_" + strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
			val := os.Getenv(key)
			if val != "" {
				if serr := fs.Set(f.Name, val); serr != nil {
					err = fmt.Errorf("invalid value %q for %s: %v", val, key, serr)
				}
			}
		}
	})
	return err
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if healthClient == nil {
		connectToRemote()
	}
	if healthClient == nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "NOT_HEALTHY")
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), timeoutDur)
	resp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{Service: serviceName})
	if err == nil && resp.Status == healthpb.HealthCheckResponse_SERVING {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
		return
	}
	glog.Errorf("server is not healthy err=%v response=%v", err, resp)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, "NOT_HEALTHY")
}

func main() {
	listen := flag.String("listen", "http://127.0.0.1:8080", "the grpc url to check health")
	flag.StringVar(&remoteUrl, "url", "", "the grpc url to check health")
	flag.StringVar(&serviceName, "service", "", "the name of the service")
	flag.BoolVar(&secureConnection, "secure-grpc", secureConnection, "do not connect to api service with tls")
	flag.BoolVar(&insecureSkipVerify, "insecure-skip-verify", insecureSkipVerify, "enable insecure skip verify tls config")
	flag.DurationVar(&timeoutDur, "timeout", timeoutDur, "timeout duration")
	flag.Parse()
	if err := SetFlagsFromEnv(flag.CommandLine, "GRPC_HEALTH"); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	l, err := url.Parse(*listen)
	if err != nil {
		glog.Fatalf("Unable to use --listen flag: %v", err)
	}
	_, p, err := net.SplitHostPort(l.Host)
	if err != nil {
		glog.Fatalf("Unable to parse host from --listen flag: %v", err)
	}
	connectToRemote()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthCheck)

	httpsrv := &http.Server{
		Addr:    fmt.Sprintf(":%s", p),
		Handler: mux,
	}
	glog.Infof("Starting grpc-health version=%s", Version)
	glog.Infof("Binding to %s...", httpsrv.Addr)
	glog.Fatal(httpsrv.ListenAndServe())
}
