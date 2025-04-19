package uints

type priority uint8

const (
	Low      priority = 0
	Medium   priority = 127
	High     priority = 200
	Critical priority = 255 // Maximum uint8 value
)

type portNumber uint16

const (
	HTTP    portNumber = 80
	HTTPS   portNumber = 443
	SSH     portNumber = 22
	Telnet  portNumber = 23
	SMTP    portNumber = 25
	DNS     portNumber = 53
	MaxPort portNumber = 65535 // Maximum uint16 value
)

type ipAddress uint32

const (
	Localhost      ipAddress = 0x7F000001 // 127.0.0.1
	GoogleDNS      ipAddress = 0x08080808 // 8.8.8.8
	Broadcast      ipAddress = 0xFFFFFFFF // 255.255.255.255
	Class_A_Subnet ipAddress = 0xFF000000 // 255.0.0.0
	Class_C_Subnet ipAddress = 0xFFFFFF00 // 255.255.255.0
)

type filesize uint64

const (
	Byte     filesize = 1
	Kilobyte filesize = 1024
	Megabyte filesize = 1024 * 1024
	Gigabyte filesize = 1024 * 1024 * 1024
	Terabyte filesize = 1024 * 1024 * 1024 * 1024
	Petabyte filesize = 1024 * 1024 * 1024 * 1024 * 1024
	// Max uint64 value would be 18,446,744,073,709,551,615
	MaxFilesize filesize = 18446744073709551615
)
