package infra

import (
	"internal/app"
	"slices"
	"testing"
)

func TestPortFromString(t *testing.T) {
	data := []struct {
		in   string
		port int
	}{
		{in: ""}, {in: "0x80"}, {in: "-1"}, {in: "65536"},
		{"80", 80}, {"1024", 1024},
	}

	for _, tt := range data {
		t.Run(
			tt.in, func(t *testing.T) {
				var dst int = -100

				err := portCI{&dst}.fromString(tt.in)

				switch err {
				case nil:
					if tt.port != dst {
						t.Fatal()
					}
				case hostPortNotMatch:
					if tt.port != 0 || dst != -100 {
						t.Fatal()
					}
				default:
					t.Fatal()
				}

			},
		)
	}
}

func TestColonHostPortFromString(t *testing.T) {
	data := []struct {
		in   string
		host string
		port int
	}{
		{in: ""}, {in: ":80"}, {in: "host:"},
		{in: "1bad.xxx:80"}, {in: "_host:80"},
		{in: "host:65536"}, {in: "host:0"}, {in: "host:0x64"},

		{"ho_st-1.2:80", "ho_st-1.2", 80},
	}

	for _, tt := range data {
		t.Run(
			tt.in, func(t *testing.T) {
				dst := app.HostColonPort{Host: "1234", Port: -100}

				err := hostColonPortCI{&dst}.fromString(tt.in)

				switch err {
				case nil:
					if tt.host != dst.Host ||
						tt.port != dst.Port {
						t.Fatal()
					}
				case hostPortNotMatch:
					if tt.host != "" || dst.Host != "1234" ||
						dst.Port != -100 {
						t.Fatal()
					}
				default:
					t.Fatal()
				}
			},
		)
	}
}

func TestHostColonPortArrayFromString(t *testing.T) {
	data := []struct {
		in  string
		err bool
		out []app.HostColonPort
	}{
		{in: "host:80,", err: true}, {in: "host:", err: true},
		{in: "host:80,1bad.xxx:80", err: true},
		{in: "", err: false, out: []app.HostColonPort{}},
		{
			in: "host:80", err: false,
			out: []app.HostColonPort{{Host: "host", Port: 80}},
		},
		{
			in: "host:80,ho_st-1.2:8080,ho-st-3.4:443", err: false,
			out: []app.HostColonPort{
				{Host: "host", Port: 80},
				{Host: "ho_st-1.2", Port: 8080},
				{Host: "ho-st-3.4", Port: 443},
			},
		},
	}

	for _, tt := range data {
		t.Run(
			tt.in, func(t *testing.T) {
				dst := []app.HostColonPort{
					{Host: "1234", Port: -100},
				}

				err := hostColonPortArrayCI{&dst}.fromString(tt.in)

				switch err {
				case nil:
					if !slices.Equal(dst, tt.out) {
						t.Fatal()
					}
				case hostPortNotMatch:
					if !tt.err || len(dst) != 1 ||
						dst[0].Host != "1234" ||
						dst[0].Port != -100 {
						t.Fatal(tt.err)
					}
				default:
					t.Fatal()
				}
			},
		)
	}
}

func TestUpdateConfig(t *testing.T) {
	port := -100
	hostPort := app.HostColonPort{
		Host: "1234",
		Port: -100,
	}

	items := map[string]configItem{
		"port":      portCI{&port},
		"host.port": hostColonPortCI{&hostPort},
	}

	if updateConfig([]byte("root.port"), []byte("80"), items) != nil {
		t.Error()
	}

	if port != 80 {
		t.Error()
	}

	if updateConfig([]byte("port"), []byte("8080"), items) != nil {
		t.Error()
	}

	if port != 80 {
		t.Error()
	}

	if updateConfig([]byte("root.host.port"), []byte("host:80"), items) != nil {
		t.Error()
	}

	if hostPort.Host != "host" || hostPort.Port != 80 {
		t.Error()
	}

	if updateConfig([]byte("root.host.port"), []byte("1host:80"), items) == nil {
		t.Error()
	}

	if hostPort.Host != "host" || hostPort.Port != 80 {
		t.Error()
	}
}
