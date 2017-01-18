package log

type F map[string]interface{}

type Log interface {
	KV(k string, v interface{}) Log
	KVs(F) Log
	Err(error) Log

	Info(string)
	Error(string)
	Warn(string)
	Fatal(string)
	Panic(string)
}

func Nop() Log {
	return nop{}
}

type nop struct{}

func (n nop) KV(k string, v interface{}) Log { return n }
func (n nop) KVs(F) Log                      { return n }
func (n nop) Err(error) Log                  { return n }
func (nop) Info(string)                      {}
func (nop) Error(string)                     {}
func (nop) Warn(string)                      {}
func (nop) Fatal(string)                     {}
func (nop) Panic(string)                     {}
