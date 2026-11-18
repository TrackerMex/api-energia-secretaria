package retranslator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"net"
	"testing"
	"time"

	"wialon-sener-pilot/internal/domain"
)

// Public Wialon Retranslator 1.0 manual, section "Example of Packet Parsing".
// The timestamp in the literal differs from the timestamp in the example's prose.
const publicExample = "74000000333533393736303133343435343835004B0BFB70000000030BBB0000" +
	"00270102706F73696E666F00A027AFDF5D9848403AC7253383DD4B4000000000" +
	"00805A40003601460B0BBB0000001200047077725F657874002B8716D9CE973B" +
	"400BBB00000011010361766C5F696E707574730000000001"

func example(t *testing.T) []byte {
	t.Helper()
	data, err := hex.DecodeString(publicExample)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func listener(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln
}

func client(t *testing.T, ln net.Listener) net.Conn {
	t.Helper()
	c, err := net.DialTimeout("tcp", ln.Addr().String(), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	if err := c.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	return c
}

func completion(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("TCP server did not stop")
		return nil
	}
}

// A third-party receiver consumes the one-way stream before closing its socket.
func finishStream(t *testing.T, c net.Conn) {
	t.Helper()
	if err := c.(*net.TCPConn).CloseWrite(); err != nil {
		t.Fatal(err)
	}
	reply, err := io.ReadAll(c)
	if err != nil || len(reply) != 0 {
		t.Fatalf("unexpected server reply %x: %v", reply, err)
	}
}

func TestR2OneWaySenderClosesWithoutReading(t *testing.T) {
	source, receiver := net.Pipe()
	defer source.Close()
	if err := source.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	count := 0
	done := make(chan error, 1)
	go func() {
		done <- receive(context.Background(), receiver, func(context.Context, domain.GPSFix) error { count++; return nil })
	}()
	data := example(t)
	_, err := source.Write(bytes.Repeat(data, 3))
	source.Close()
	if stopErr := completion(t, done); stopErr != nil {
		t.Fatal(stopErr)
	}
	if err != nil || count != 3 {
		t.Fatalf("sender delivered %d of 3 frames: %v", count, err)
	}
}

func TestR2OfficialPacket(t *testing.T) {
	b := example(t)
	fix, present, err := decodeFrame(b)
	if err != nil || !present {
		t.Fatalf("position=%t: %v", present, err)
	}
	if fix.UID != "353976013445485" || !fix.GPSTime.Equal(time.Unix(0x4b0bfb70, 0)) || fix.GPSTime.Location() != time.UTC {
		t.Fatalf("unexpected unit or timestamp: %+v", fix)
	}
	if math.Abs(fix.Longitude-49.1903648) > 1e-7 || math.Abs(fix.Latitude-55.7305664) > 1e-7 || fix.Speed != 54 || fix.Course != 326 {
		t.Fatalf("unexpected navigation values: %+v", fix)
	}
	if fix.Fingerprint != sha256.Sum256(b) {
		t.Fatal("must fingerprint the entire wire frame")
	}
}

func TestR2FragmentedAndCoalesced(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ln := listener(t)
	received := make(chan domain.GPSFix, 2)
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, ln, func(_ context.Context, fix domain.GPSFix) error { received <- fix; return nil })
	}()
	c := client(t, ln)
	first := example(t)
	second := bytes.Clone(first)
	second[len(second)-1] ^= 1 // different sensor, same GPS second
	for _, chunk := range [][]byte{first[:1], first[1:7], append(bytes.Clone(first[7:]), second...)} {
		if _, err := c.Write(chunk); err != nil {
			t.Fatal(err)
		}
	}
	finishStream(t, c)
	if len(received) != 2 {
		t.Fatalf("received %d frames", len(received))
	}
	a, b := <-received, <-received
	if a.Fingerprint == b.Fingerprint || !a.GPSTime.Equal(b.GPSTime) {
		t.Fatal("lost wire identity of same-second frames")
	}
	cancel()
	if err := completion(t, done); err != nil {
		t.Fatal(err)
	}
}

func TestR5InvalidAndNonPositionFrames(t *testing.T) {
	valid := example(t)
	posOffset := bytes.Index(valid, []byte("posinfo\x00")) + len("posinfo\x00")
	for name, alter := range map[string]func([]byte){
		"longitude":          func(f []byte) { binary.LittleEndian.PutUint64(f[posOffset:], math.Float64bits(181)) },
		"latitude":           func(f []byte) { binary.LittleEndian.PutUint64(f[posOffset+8:], math.Float64bits(-91)) },
		"nan":                func(f []byte) { binary.LittleEndian.PutUint64(f[posOffset:], math.Float64bits(math.NaN())) },
		"course":             func(f []byte) { binary.BigEndian.PutUint16(f[posOffset+26:], 360) },
		"location bit clear": func(f []byte) { binary.BigEndian.PutUint32(f[24:], 2) },
		"block boundary":     func(f []byte) { binary.BigEndian.PutUint32(f[30:], 0xffff) },
	} {
		t.Run(name, func(t *testing.T) {
			frame := bytes.Clone(valid)
			alter(frame)
			if _, present, err := decodeFrame(frame); err == nil || present {
				t.Fatal("bad frame produced a position")
			}
		})
	}
	// Keep the ordinary sensor blocks, remove the position block and its mask bit.
	ordinary := append(bytes.Clone(valid[:28]), valid[73:]...)
	binary.LittleEndian.PutUint32(ordinary, uint32(len(ordinary)-4))
	binary.BigEndian.PutUint32(ordinary[24:], 2)
	if _, present, err := decodeFrame(ordinary); err != nil || present {
		t.Fatalf("sensor frame: position=%t, err=%v", present, err)
	}
}

func TestR5PeerIsolationAndIdleShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ln := listener(t)
	done := make(chan error, 1)
	saved := make(chan domain.GPSFix, 4)
	go func() {
		done <- Serve(ctx, ln, func(_ context.Context, p domain.GPSFix) error { saved <- p; return nil })
	}()
	full := example(t)
	for _, bad := range [][]byte{{0xff, 0xff, 0xff, 0xff}, full[:13]} {
		c := client(t, ln)
		if _, err := c.Write(bad); err != nil {
			t.Fatal(err)
		}
		c.Close()
	}
	c := client(t, ln)
	if _, err := c.Write(full); err != nil {
		t.Fatal(err)
	}
	finishStream(t, c)
	idle := client(t, ln)
	if _, err := idle.Write(full[:2]); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := completion(t, done); err != nil {
		t.Fatal(err)
	}
	if len(saved) != 1 {
		t.Fatalf("expected one committed frame; got %d", len(saved))
	}
	if _, err := idle.Read(make([]byte, 1)); err == nil {
		t.Fatal("idle connection survived shutdown")
	}
}

func TestR5StorageFailureIsReturned(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ln := listener(t)
	done := make(chan error, 1)
	storeErr := errors.New("storage unavailable")
	go func() { done <- Serve(ctx, ln, func(context.Context, domain.GPSFix) error { return storeErr }) }()
	c := client(t, ln)
	if _, err := c.Write(example(t)); err != nil {
		t.Fatal(err)
	}
	if err := completion(t, done); !errors.Is(err, storeErr) {
		t.Fatalf("lost storage error: %v", err)
	}
	if n, _ := c.Read(make([]byte, 1)); n != 0 {
		t.Fatal("third-party receiver should emit no payload")
	}
}
