package wakeonlan

import (
"bytes"
"testing"
)

// TestWriteMAC tests the WriteMAC method with various MAC address formats.
func TestWriteMAC(t *testing.T) {
tests := []struct {
name    string
macAddr string
wantErr bool
errMsg  string
}{
{
name:    "valid MAC with colons",
macAddr: "00:11:22:33:44:55",
wantErr: false,
},
{
name:    "valid MAC with dashes",
macAddr: "00-11-22-33-44-55",
wantErr: false,
},
{
name:    "valid MAC uppercase",
macAddr: "AA:BB:CC:DD:EE:FF",
wantErr: false,
},
{
name:    "invalid MAC - too short",
macAddr: "00:11:22:33:44",
wantErr: true,
errMsg:  "invalid MAC address",
},
{
name:    "invalid MAC - empty string",
macAddr: "",
wantErr: true,
errMsg:  "invalid MAC address",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
p := &Packet{}
p.writeHeader()
err := p.WriteMAC(tt.macAddr)

if (err != nil) != tt.wantErr {
t.Errorf("WriteMAC() error = %v, wantErr %v", err, tt.wantErr)
return
}

if !tt.wantErr && p.payload.Len() != 102 {
t.Errorf("WriteMAC() payload size = %d, want 102", p.payload.Len())
}
})
}
}

// TestWriteMACPayloadContent verifies the packet contains the correct MAC address repeated.
func TestWriteMACPayloadContent(t *testing.T) {
p := &Packet{}
p.writeHeader()

macAddr := "aa:bb:cc:dd:ee:ff"
err := p.WriteMAC(macAddr)
if err != nil {
t.Fatalf("WriteMAC() unexpected error: %v", err)
}

payload := p.payload.Bytes()

// First 6 bytes should be 0xFF (header)
for i := 0; i < 6; i++ {
if payload[i] != 0xFF {
t.Errorf("Header byte %d = 0x%02X, want 0xFF", i, payload[i])
}
}

// MAC address pattern: aa, bb, cc, dd, ee, ff
expected := []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}

// Verify MAC is repeated 16 times after header
for rep := 0; rep < 16; rep++ {
offset := 6 + rep*6
for j := 0; j < 6; j++ {
if payload[offset+j] != expected[j] {
t.Errorf("MAC repetition %d, byte %d = 0x%02X, want 0x%02X",
rep, j, payload[offset+j], expected[j])
}
}
}
}

// TestGetAddr tests IP address parsing with various inputs.
func TestGetAddr(t *testing.T) {
tests := []struct {
name    string
addr    string
want    string
wantErr bool
}{
{
name:    "empty address returns broadcast",
addr:    "",
want:    "255.255.255.255",
wantErr: false,
},
{
name:    "valid broadcast address",
addr:    "255.255.255.255",
want:    "255.255.255.255",
wantErr: false,
},
{
name:    "valid private IP",
addr:    "192.168.1.100",
want:    "192.168.1.100",
wantErr: false,
},
{
name:    "invalid IP - malformed",
addr:    "256.256.256.256",
wantErr: true,
},
{
name:    "invalid IP - text",
addr:    "invalid.ip.address",
wantErr: true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got, err := getRAddr(tt.addr)

if (err != nil) != tt.wantErr {
t.Errorf("getRAddr() error = %v, wantErr %v", err, tt.wantErr)
return
}

if !tt.wantErr && got.String() != tt.want {
t.Errorf("getRAddr() = %v, want %v", got.String(), tt.want)
}
})
}
}

// TestGetPort tests port parsing and validation.
func TestGetPort(t *testing.T) {
tests := []struct {
name    string
port    string
want    int
wantErr bool
}{
{
name:    "empty port returns default (9)",
port:    "",
want:    9,
wantErr: false,
},
{
name:    "valid port 0 (any)",
port:    "0",
want:    0,
wantErr: false,
},
{
name:    "valid port 7 (echo)",
port:    "7",
want:    7,
wantErr: false,
},
{
name:    "valid port 9 (discard)",
port:    "9",
want:    9,
wantErr: false,
},
{
name:    "invalid port - not accepted",
port:    "80",
wantErr: true,
},
{
name:    "invalid port - non-numeric",
port:    "abc",
wantErr: true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got, err := getPort(tt.port)

if (err != nil) != tt.wantErr {
t.Errorf("getPort() error = %v, wantErr %v", err, tt.wantErr)
return
}

if !tt.wantErr && got != tt.want {
t.Errorf("getPort() = %d, want %d", got, tt.want)
}
})
}
}

// TestSendMagicValidation tests SendMagic with invalid parameters.
func TestSendMagicValidation(t *testing.T) {
tests := []struct {
name    string
macAddr string
passwd  string
wantErr bool
}{
{
name:    "password not supported",
macAddr: "00:11:22:33:44:55",
passwd:  "password",
wantErr: true,
},
{
name:    "invalid MAC address",
macAddr: "invalid",
passwd:  "",
wantErr: true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
err := SendMagic(tt.macAddr, tt.passwd, "255.255.255.255", "9")

if (err != nil) != tt.wantErr {
t.Errorf("SendMagic() error = %v, wantErr %v", err, tt.wantErr)
}
})
}
}

// TestPacketConstruction verifies the packet structure is built correctly.
func TestPacketConstruction(t *testing.T) {
p := &Packet{}

if p.payload.Len() != 0 {
t.Errorf("New packet should be empty, got length %d", p.payload.Len())
}

p.writeHeader()
if p.payload.Len() != 6 {
t.Errorf("After header, packet size = %d, want 6", p.payload.Len())
}

err := p.WriteMAC("00:11:22:33:44:55")
if err != nil {
t.Fatalf("WriteMAC() unexpected error: %v", err)
}

if p.payload.Len() != 102 {
t.Errorf("After MAC, packet size = %d, want 102", p.payload.Len())
}
}

// TestWriteHeaderContent verifies header bytes are exactly 0xFF.
func TestWriteHeaderContent(t *testing.T) {
p := &Packet{}
p.writeHeader()

payload := p.payload.Bytes()
if len(payload) != 6 {
t.Fatalf("Header size = %d, want 6", len(payload))
}

for i, b := range payload {
if b != 0xFF {
t.Errorf("Header byte %d = 0x%02X, want 0xFF", i, b)
}
}
}

// BenchmarkWriteMAC benchmarks the WriteMAC operation.
func BenchmarkWriteMAC(b *testing.B) {
p := &Packet{}

for i := 0; i < b.N; i++ {
p.payload.Reset()
p.writeHeader()
_ = p.WriteMAC("00:11:22:33:44:55")
}
}

// BenchmarkGetAddr benchmarks IP address parsing.
func BenchmarkGetAddr(b *testing.B) {
for i := 0; i < b.N; i++ {
_, _ = getRAddr("192.168.1.100")
}
}

// BenchmarkGetPort benchmarks port parsing.
func BenchmarkGetPort(b *testing.B) {
for i := 0; i < b.N; i++ {
_, _ = getPort("9")
}
}

// TestMACFormatVariations tests different valid MAC address formats.
func TestMACFormatVariations(t *testing.T) {
validMACFormats := []string{
"00:00:00:00:00:00",
"ff:ff:ff:ff:ff:ff",
"01:23:45:67:89:ab",
"01-23-45-67-89-ab",
"AA:BB:CC:DD:EE:FF",
}

for _, macAddr := range validMACFormats {
t.Run(macAddr, func(t *testing.T) {
p := &Packet{}
p.writeHeader()
err := p.WriteMAC(macAddr)

if err != nil {
t.Errorf("WriteMAC(%q) unexpected error: %v", macAddr, err)
}

if p.payload.Len() != 102 {
t.Errorf("WriteMAC(%q) payload size = %d, want 102", macAddr, p.payload.Len())
}
})
}
}

// TestIPBroadcastFormats tests various broadcast IP configurations.
func TestIPBroadcastFormats(t *testing.T) {
tests := []struct {
ip   string
want string
}{
{"", "255.255.255.255"},
{"255.255.255.255", "255.255.255.255"},
{"192.168.0.255", "192.168.0.255"},
{"10.0.0.255", "10.0.0.255"},
}

for _, tt := range tests {
t.Run(tt.ip, func(t *testing.T) {
got, err := getRAddr(tt.ip)
if err != nil {
t.Errorf("getRAddr(%q) unexpected error: %v", tt.ip, err)
return
}

if got.String() != tt.want {
t.Errorf("getRAddr(%q) = %v, want %v", tt.ip, got.String(), tt.want)
}
})
}
}

// TestPortDefaults verifies empty port defaults to 9.
func TestPortDefaults(t *testing.T) {
got, err := getPort("")
if err != nil {
t.Fatalf("getPort(\"\") unexpected error: %v", err)
}

if got != 9 {
t.Errorf("getPort(\"\") = %d, want 9 (default)", got)
}
}

// TestErrorWrapping verifies errors maintain context through wrapping.
func TestErrorWrapping(t *testing.T) {
err := SendMagic("invalid-mac", "", "255.255.255.255", "9")
if err == nil {
t.Fatal("SendMagic with invalid MAC should return error")
}

errStr := err.Error()
if !bytes.Contains([]byte(errStr), []byte("MAC")) {
t.Errorf("Error should mention MAC: %v", err)
}
}
