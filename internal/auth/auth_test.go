package auth

import "testing"

func TestAuthenticate(t *testing.T) {
	a := New("tok1:t1:sub1,tok2:t2:sub2")
	claims, err := a.Authenticate("Bearer tok1")
	if err != nil {
		t.Fatalf("unexpected auth error: %v", err)
	}
	if claims.TenantID != "t1" || claims.Subject != "sub1" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if _, err := a.Authenticate("Bearer nope"); err == nil {
		t.Fatalf("expected invalid token error")
	}
}
