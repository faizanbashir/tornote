package tornote

const (
	TestDSN    = "postgres://postgres:postgres@localhost/testdb"
	TestPort   = 31337
	TestSecret = "0123456789"
)

func stubServer() *server {
	s := NewServer(TestPort, TestDSN, TestSecret)
	s.Init()
	return s
}

//func TestNewServer(t *testing.T) {
//	s := stubServer()
//	if s.DSN != TestDSN && s.Port != TestPort {
//		t.Fatal("can not initialize server")
//	}
//}
//
//func TestServer_Run(t *testing.T) {
//	s := stubServer()
//
//}
