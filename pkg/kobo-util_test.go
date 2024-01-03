package pkg

import "testing"

func TestGetDevice(t *testing.T) {
	type args struct {
		fn string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "",
			args: args{
				fn: "./fixtures/xxx",
			},
			want:    "N418180050132",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDevice(tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDevice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getDevice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetModel(t *testing.T) {
	type args struct {
		model string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Libra",
			args: args{
				model: "N418180050132",
			},
			want: "Kobo Libra 2",
		},
		{
			name: "Clara",
			args: args{
				model: "N50629C039232",
			},
			want: "Kobo Clara 2E",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getModel(tt.args.model); got != tt.want {
				t.Errorf("deviceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
