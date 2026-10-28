package hl7_test

import (
	hl7 "github.com/synkwise/go-hl7-parser"
	"os"
	"strings"
	"testing"
)

func readFile(fname string) ([]byte, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make([]byte, 1024)
	if _, err = file.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

func TestMessage(t *testing.T) {
	data, err := readFile("./testdata/msg.hl7")
	if err != nil {
		t.Fatal(err)
	}

	msg := &hl7.Message{Value: data}
	msg.Parse()
	if err != nil {
		t.Error(err)
	}
	if len(msg.Segments) != 5 {
		t.Errorf("Expected 5 segments got %d\n", len(msg.Segments))
	}

	data, err = readFile("./testdata/msg2.hl7")
	if err != nil {
		t.Fatal(err)
	}
	msg = &hl7.Message{Value: data}
	msg.Parse()
	if err != nil {
		t.Error(err)
	}
	if len(msg.Segments) != 5 {
		t.Errorf("Expected 5 segments got %d\n", len(msg.Segments))
	}

	data, err = readFile("./testdata/msg3.hl7")
	if err != nil {
		t.Fatal(err)
	}
	msg = &hl7.Message{Value: data}
	msg.Parse()
	if err != nil {
		t.Error(err)
	}
	if len(msg.Segments) != 9 {
		t.Errorf("Expected 9 segments got %d\n", len(msg.Segments))
	}

	data, err = readFile("./testdata/msg4.hl7")
	if err != nil {
		t.Fatal(err)
	}
	msg = &hl7.Message{Value: data}
	msg.Parse()
	if err != nil {
		t.Error(err)
	}
	if len(msg.Segments) != 9 {
		t.Errorf("Expected 9 segments got %d\n", len(msg.Segments))
	}
}

func TestMsgUnmarshal(t *testing.T) {
	fname := "./testdata/msg.hl7"
	file, err := os.Open(fname)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	msgs, err := hl7.NewDecoder(file).Messages()
	if err != nil {
		t.Fatal(err)
	}
	st := my7{}
	msgs[0].Unmarshal(&st)

	if st.FirstName != "John" {
		t.Errorf("Expected John got %s\n", st.FirstName)
	}
	if st.LastName != "Jones" {
		t.Errorf("Expected Jones got %s\n", st.LastName)
	}
}

func sizedMessage(size int) []byte {
	head := "MSH|^~\\&|A|B|C|D|20260101||ADT^A01|1|P|2.3\rPID|1||"
	pad := strings.Repeat("X", size-len(head)-1)
	return []byte(head + pad + "\r")
}

func TestParseFullBuffer(t *testing.T) {
	for _, size := range []int{175, 176, 288, 320, 1024} {
		data := sizedMessage(size)
		msg := hl7.NewMessage(data[:size:size])
		if err := msg.Parse(); err != nil {
			t.Errorf("size %d: unexpected error: %v", size, err)
			continue
		}
		if len(msg.Segments) != 2 {
			t.Errorf("size %d: expected 2 segments got %d", size, len(msg.Segments))
		}
	}
}

func TestParseLastSegmentWithoutTerminator(t *testing.T) {
	src := "MSH|^~\\&|A|B|C|D|20260101||ADT^A01|1|P|2.3\rPID|1||SMITH"
	data := make([]byte, len(src), len(src)+4)
	copy(data, src)
	data[:cap(data)][len(data)] = '#'

	msg := hl7.NewMessage(data)
	if err := msg.Parse(); err != nil {
		t.Fatal(err)
	}
	if len(msg.Segments) != 2 {
		t.Fatalf("expected 2 segments got %d", len(msg.Segments))
	}
	if got := string(msg.Segments[1].Value); got != "PID|1||SMITH" {
		t.Errorf("expected last segment %q got %q", "PID|1||SMITH", got)
	}
}

func TestParseMultiByteCharacters(t *testing.T) {
	src := "MSH|^~\\&|A|B|C|D|20260101||ADT^A01|1|P|2.3\rPID|1||MÜLLER&JÖ^ANNA|X\rPV1|1|I\r"
	msg := hl7.NewMessage([]byte(src))
	if err := msg.Parse(); err != nil {
		t.Fatal(err)
	}
	if len(msg.Segments) != 3 {
		t.Fatalf("expected 3 segments got %d", len(msg.Segments))
	}
	if got := string(msg.Segments[1].Value); got != "PID|1||MÜLLER&JÖ^ANNA|X\r" {
		t.Errorf("unexpected PID segment %q", got)
	}
	if got := string(msg.Segments[2].Value); got != "PV1|1|I\r" {
		t.Errorf("unexpected PV1 segment %q", got)
	}

	name := msg.Segments[1].Fields[3]
	if got := string(name.Value); got != "MÜLLER&JÖ^ANNA" {
		t.Fatalf("unexpected name field %q", got)
	}
	if len(name.Components) != 2 {
		t.Fatalf("expected 2 components got %d", len(name.Components))
	}
	if got := string(name.Components[1].Value); got != "ANNA" {
		t.Errorf("unexpected given name %q", got)
	}
	family := name.Components[0].SubComponents
	if len(family) != 2 || string(family[0].Value) != "MÜLLER" || string(family[1].Value) != "JÖ" {
		t.Errorf("unexpected family name subcomponents %q", family)
	}
}
