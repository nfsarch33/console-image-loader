package console_image

import (
	"testing"
)

func TestShowImg(t *testing.T) {
	type args struct {
		imagePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Show static image",
			args: args{
				imagePath: "test_image/goku.jpg",
			},
			wantErr: false,
		},
		{
			name: "Show animated GIF",
			args: args{
				imagePath: "test_image/goku.gif",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ShowImg(tt.args.imagePath); (err != nil) != tt.wantErr {
				t.Errorf("ShowImg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
