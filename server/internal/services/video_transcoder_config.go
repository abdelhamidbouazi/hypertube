package services

import (
	"log"

	"github.com/spf13/viper"
)

var VideoTranscoderConf VideoTranscoderConfig

type VideoTranscoderConfig struct {
	Encoding struct {
		Preset          string `mapstructure:"preset"`
		CRF             int    `mapstructure:"crf"`
		GOPSize         int    `mapstructure:"gop_size"`
		KeyintMin       int    `mapstructure:"keyint_min"`
		SCThreshold     int    `mapstructure:"sc_threshold"`
		AudioBitrate    string `mapstructure:"audio_bitrate"`
		AudioSampleRate int    `mapstructure:"audio_sample_rate"`
	} `mapstructure:"encoding"`

	Qualities []struct {
		Name         string `mapstructure:"name"`
		Resolution   string `mapstructure:"resolution"`
		VideoBitrate string `mapstructure:"video_bitrate"`
		MaxRate      string `mapstructure:"max_rate"`
		BufSize      string `mapstructure:"buf_size"`
		Enabled      bool   `mapstructure:"enabled"`
	} `mapstructure:"qualities"`

	Output struct {
		Directory         string `mapstructure:"directory"`
		SegmentTime       int    `mapstructure:"segment_time"`
		PlaylistType      string `mapstructure:"playlist_type"`
		UseTemporaryFiles bool   `mapstructure:"use_temporary_files"`
		DeleteOldSegments bool   `mapstructure:"delete_old_segments"`
	} `mapstructure:"output"`
}

func LoadVideoTranscoderConfig(configPath string) {
	Logger.Debug("Loading Video Transcoder Config")

	v := viper.New()
	v.SetConfigFile(configPath)

	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading video transcoder config file: %v", err)
	}

	err = v.Unmarshal(&VideoTranscoderConf)
	if err != nil {
		log.Fatalf("Unable to decode video transcoder config into struct: %v", err)
	}

	Logger.Info("Video Transcoder Config loaded successfully")
}
