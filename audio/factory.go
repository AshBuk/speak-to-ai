package audio

import (
	"fmt"

	"github.com/AshBuk/speak-to-ai/config"
)

// AudioRecorderFactory creates appropriate audio recorder instances
type AudioRecorderFactory struct {
	config *config.Config
}

// NewAudioRecorderFactory creates a new factory instance
func NewAudioRecorderFactory(config *config.Config) *AudioRecorderFactory {
	return &AudioRecorderFactory{
		config: config,
	}
}

// CreateRecorder creates a recorder based on config settings
func (f *AudioRecorderFactory) CreateRecorder() (AudioRecorder, error) {
	switch f.config.Audio.RecordingMethod {
	case "arecord":
		return NewArecordRecorder(f.config), nil
	case "ffmpeg":
		return NewFFmpegRecorder(f.config), nil
	default:
		return nil, fmt.Errorf("unsupported recording method: %s", f.config.Audio.RecordingMethod)
	}
}

// GetRecorder is a convenience function to create a recorder directly from config
func GetRecorder(config *config.Config) (AudioRecorder, error) {
	factory := NewAudioRecorderFactory(config)
	return factory.CreateRecorder()
}
