package seeds

import "testing"

func TestOrgScopedWorkerBootstrapTokenHashes(t *testing.T) {
	tests := map[string]struct {
		got  string
		want string
	}{
		"smtp": {
			got:  smtpOrgBootstrapTokenHash,
			want: "7fb1131c8efddfa50392b96c971ee3358d18a10c192eaf97e409b916558a22bc",
		},
		"template": {
			got:  templateOrgBootstrapTokenHash,
			want: "16f657f9a9a798fe31dc35dd9b601499d56a285accfde0748e1af9a035f7d5fc",
		},
		"telegram": {
			got:  telegramOrgBootstrapTokenHash,
			want: "79b8a204126030209d98fa2f53e730e3da31f891e7691e82e7c419d337b69d67",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("hash = %q, want %q", tt.got, tt.want)
			}
		})
	}
}
