package cmd

import "testing"

func TestValidateResourceId(t *testing.T) {
	type args struct {
		resourceId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid resource id",
			args:    args{resourceId: "/subscriptions/64f552d2-406c-4842-836f-6cd00c086f7e/resourceGroups/huuhkadotnet/providers/Microsoft.OperationalInsights/workspaces/huuhkadotnet-law-prod"},
			wantErr: false,
		},
		{
			name:    "valid subresource id",
			args:    args{resourceId: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/myVM/extensions/someExtension"},
			wantErr: false,
		},
		{
			name:    "invalid resource id",
			args:    args{resourceId: "invalid"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateResourceId(tt.args.resourceId); (err != nil) != tt.wantErr {
				t.Errorf("validateResourceId() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}