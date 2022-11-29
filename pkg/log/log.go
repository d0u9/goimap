package log

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_levelToColor = map[zapcore.Level]*color.Color{
		zapcore.DebugLevel:  color.New(color.FgMagenta),
		zapcore.InfoLevel:   color.New(color.FgBlue),
		zapcore.WarnLevel:   color.New(color.FgYellow),
		zapcore.ErrorLevel:  color.New(color.FgRed),
		zapcore.DPanicLevel: color.New(color.FgRed),
		zapcore.PanicLevel:  color.New(color.FgRed),
		zapcore.FatalLevel:  color.New(color.FgRed),
	}
	_unknownLevelColor = color.New(color.FgRed)

	_levelToCapitalColorString = make(map[zapcore.Level]string, len(_levelToColor))
)

func init() {
	for level, color := range _levelToColor {
		var s string
		switch level {
		case zapcore.DebugLevel:
			s = color.Sprint("[D]")
		case zapcore.InfoLevel:
			s = color.Sprint("[I]")
		case zapcore.WarnLevel:
			s = color.Sprint("[W]")
		case zapcore.ErrorLevel:
			s = color.Sprint("[E]")
		case zapcore.DPanicLevel:
			s = color.Sprint("[P]")
		case zapcore.PanicLevel:
			s = color.Sprint("[P]")
		case zapcore.FatalLevel:
			s = color.Sprint("[F]")
		}
		_levelToCapitalColorString[level] = s
	}

	if err := initLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "cannot init logger: %v", err)
		os.Exit(1)
	}
}

func initLogger() error {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = LevelEncoder
	config.EncoderConfig.EncodeTime = TimerEncoder
	config.EncoderConfig.EncodeCaller = CallerEncoder
	config.EncoderConfig.ConsoleSeparator = " "
	config.DisableStacktrace = true
	logger, err := config.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func LevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := _levelToCapitalColorString[l]
	if !ok {
		s = _unknownLevelColor.Sprint("[U]")
	}
	enc.AppendString(s)
}

func TimerEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	s := color.New(color.FgBlue).Sprintf("[%s]", t.Format("Jan _2 15:04:05"))
	enc.AppendString(s)
}

func CallerEncoder(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	s := color.New(color.FgCyan).Sprintf("[%s]", c.TrimmedPath())
	enc.AppendString(s)
}

type WrapLogger struct {
	*zap.SugaredLogger
}

func NewWrapLogger() *WrapLogger {
	return &WrapLogger{
		SugaredLogger: zap.S(),
	}
}

func (l *WrapLogger) Printf(msg string, args ...any) {
	l.Errorf(msg, args...)
}

func (l *WrapLogger) Println(args ...any) {
	l.Errorln(args...)
}
