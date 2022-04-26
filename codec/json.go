package codec

import (
	"bufio"
	"encoding/json"
	"github.com/DurantVivado/GrasureOL/xlog"
	"io"
)

type JsonCodec struct{
	conn io.ReadWriteCloser
	buf *bufio.Writer
	dec *json.Decoder
	enc *json.Encoder
}

func (j *JsonCodec) Close() error {
	return j.Close()
}

func (j *JsonCodec) ReadHeader(header *Header) error {
	return j.dec.Decode(header)
}

func (j *JsonCodec) ReadBody(body interface{}) error {
	return j.dec.Decode(body)
}

func (j *JsonCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = j.buf.Flush()
		if err != nil{
			_ = j.Close()
		}
	}()
	if err := j.enc.Encode(header); err != nil{
		xlog.Errorln("rpc codec: json error encoding header", err)
		return err
	}
	if err := j.enc.Encode(body); err != nil{
		xlog.Errorln("rpc codec: json error encoding body:", err)
		return err
	}
	return nil
}

var _Codec = (*JsonCodec) (nil)

func NewJsonCodec( conn io.ReadWriteCloser) Codec{
	buf := bufio.NewWriter(conn)
	return &JsonCodec{
		conn: conn,
		buf:  buf,
		dec:  json.NewDecoder(conn),
		enc:  json.NewEncoder(buf),
	}
}
