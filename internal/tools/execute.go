package tools

import (
	"context"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func ExecuteToolCalls(ctx context.Context, allowedTools []einotool.BaseTool, assistantMessage *schema.Message) ([]*schema.Message, error) {
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools:               allowedTools,
		ExecuteSequentially: true,
	})
	if err != nil {
		return nil, err
	}

	return toolsNode.Invoke(ctx, assistantMessage)
}
