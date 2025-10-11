package mock

import (
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/mock"
)

func CaptureAllOutput(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	pterm.SetDefaultOutput(w)

	printers := []*pterm.PrefixPrinter{
		&pterm.Info,
		&pterm.Error,
		&pterm.Warning,
		&pterm.Success,
		&pterm.Debug,
		&pterm.Fatal,
	}
	oldWriters := make([]io.Writer, len(printers))
	for i, p := range printers {
		oldWriters[i] = p.Writer
		p.Writer = w
	}

	fn()

	_ = w.Close()
	os.Stdout = oldStdout
	for i, p := range printers {
		p.Writer = oldWriters[i]
	}

	out, _ := io.ReadAll(r)
	return string(out)
}

type TestLogger struct {
	Data string
}

func (l *TestLogger) LogMessage(data string) { l.Data = data }
func (l *TestLogger) GetLogContext() string  { return "" }

type MockSTSClient struct {
	Output *sts.GetCallerIdentityOutput
	Err    error
}

func (m *MockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return m.Output, m.Err
}

type MockInvoker struct {
	mock.Mock
}

func (m *MockInvoker) InvokeModel(ctx context.Context, input *bedrockruntime.InvokeModelInput) (*bedrockruntime.InvokeModelOutput, error) {
	args := m.Called(ctx, input)
	if out := args.Get(0); out != nil {
		return out.(*bedrockruntime.InvokeModelOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInvoker) ListFoundationModels(ctx context.Context, input *bedrock.ListFoundationModelsInput) (*bedrock.ListFoundationModelsOutput, error) {
	args := m.Called(ctx, input)
	if out := args.Get(0); out != nil {
		return out.(*bedrock.ListFoundationModelsOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInvoker) ConverseModel(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
	args := m.Called(ctx, input)
	if out := args.Get(0); out != nil {
		return out.(*bedrockruntime.ConverseOutput), args.Error(1)
	}
	return nil, args.Error(1)
}
