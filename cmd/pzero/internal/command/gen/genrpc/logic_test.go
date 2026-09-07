package genrpc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChangeLogicTypesNormalizesCanonicalImportAlias(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "request_refund.go")
	contents := `package paymentservicelogic

import paymentv1 "github.com/example/contracts/gen/payment/v1"

type RequestRefund struct{}

func (l *RequestRefund) RequestRefund(in *paymentv1.OldRequest) (*paymentv1.OldResponse, error) {
	return nil, nil
}
`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	jr := &PzeroRpc{Module: "github.com/example/payment-svc"}
	err := jr.changeLogicTypes(LogicFile{
		Path:             filepath.Join(dir, "request_refund_logic.go"),
		Package:          "payment",
		GoPackage:        "github.com/example/contracts/gen/payment/v1;payment",
		Service:          "PaymentService",
		Rpc:              "RequestRefund",
		RequestTypeName:  "RequestRefundRequest",
		ResponseTypeName: "RequestRefundResponse",
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "*payment.RequestRefundRequest") ||
		!strings.Contains(string(got), "*payment.RequestRefundResponse") {
		t.Fatalf("canonical import alias was not normalized:\n%s", got)
	}
	if strings.Contains(string(got), "paymentv1.") {
		t.Fatalf("stale protobuf package qualifier leaked into logic:\n%s", got)
	}
}
