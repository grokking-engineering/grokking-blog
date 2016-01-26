package logs

import (
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

type M map[string]interface{}

type Logger struct {
	logger *logrus.Logger
}

type Entry struct {
	*logrus.Entry
}

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stderr)
}

func New(prefix string) Logger {
	logger := logrus.New()
	return Logger{logger}
}

// 2015-06-10 20:10:08.123456
func getTime() string {
	var buf [30]byte
	b := buf[:0]
	t := time.Now()
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	nsec := t.Nanosecond()

	itoa(&b, year, 4)
	b = append(b, '-')
	itoa(&b, int(month), 2)
	b = append(b, '-')
	itoa(&b, day, 2)
	b = append(b, ' ')
	itoa(&b, hour, 2)
	b = append(b, ':')
	itoa(&b, min, 2)
	b = append(b, ':')
	itoa(&b, sec, 2)
	b = append(b, '.')
	itoa(&b, nsec/1e3, 6)

	return string(b)
}

// Taken from stdlib "logrus".
//
// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func trimFile(path string) string {
	index := strings.Index(path, "/src/")
	if index < 0 {
		return path
	}
	return path[index+5:]
}

func (l Logger) WithError(err error) Entry {
	_, file, line, _ := runtime.Caller(2)
	return l.WithFields(logrus.Fields{
		"*TIME": getTime(),
		".FILE": trimFile(file),
		".LINE": line,
		"error": err,
	})
}

func (l Logger) MaybePanic(err error) Entry {
	_, file, line, _ := runtime.Caller(2)
	debug.PrintStack()
	return l.WithFields(logrus.Fields{
		"*TIME": getTime(),
		".FILE": trimFile(file),
		".LINE": line,
		"error": err,
	})
}

func (l Logger) WithFields(fields map[string]interface{}) Entry {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	_, file, line, _ := runtime.Caller(2)
	fields["*TIME"] = getTime()
	fields[".FILE"] = trimFile(file)
	fields[".LINE"] = line
	return Entry{
		l.logger.WithFields(logrus.Fields(fields)),
	}
}

func (l Logger) Fatal(msg interface{}) {
	debug.PrintStack()
	l.logger.Fatal("Fatal: ", msg)
}

func (l Logger) Error(msg ...interface{}) {
	l.logger.Error(msg...)
}

func (l Logger) Println(msg ...interface{}) {
	l.logger.Println(msg...)
}

func (l Logger) Printf(format string, msg ...interface{}) {
	l.logger.Printf(format, msg...)
}

func (e Entry) WithFields(fields map[string]interface{}) Entry {
	return Entry{
		e.Entry.WithFields(logrus.Fields(fields)),
	}
}

func (e Entry) Fatal(msg interface{}) {
	debug.PrintStack()
	e.Entry.Fatal("Fatal: ", msg)
}
