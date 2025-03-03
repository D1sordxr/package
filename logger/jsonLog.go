package logger

import (
	"bytes"
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type PrettyJSONEncoder struct {
	zapcore.Encoder
}

func (e *PrettyJSONEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) ([]byte, error) {
	// Создаём обычный JSON-объект
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:      "time",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeTime:   zapcore.TimeEncoderOfLayout("[15:05:05.000]"),
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	// Кодируем как JSON
	buf, err := jsonEncoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	// Форматируем JSON красиво (многострочно)
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, buf.Bytes(), "", "  ")
	if err != nil {
		return nil, err
	}

	// Возвращаем отформатированный JSON
	return prettyJSON.Bytes(), nil
}

// NewZapLogger создает логгер с многострочным JSON-форматом
func NewZapLogger() *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:      "time",
			LevelKey:     "level",
			MessageKey:   "message",
			CallerKey:    "caller",
			EncodeTime:   zapcore.TimeEncoderOfLayout("[15:05:05.000]"),
			EncodeLevel:  zapcore.CapitalColorLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		}),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	)

	return zap.New(core, zap.AddCaller())
}
