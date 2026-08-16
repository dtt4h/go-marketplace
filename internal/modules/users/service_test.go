package users

import "testing"

func TestValidateProfileAndStoreInputs(t *testing.T) {
	t.Run("rejects empty username", func(t *testing.T) {
		username := "   "
		if err := validateUsername(&username); err != ErrUsernameRequired {
			t.Fatalf("expected ErrUsernameRequired, got %v", err)
		}
	})

	t.Run("rejects empty store name", func(t *testing.T) {
		if err := validateStoreName("   "); err != ErrStoreNameRequired {
			t.Fatalf("expected ErrStoreNameRequired, got %v", err)
		}
	})
}
