package v1

const (
	ModelProviderFinalizer         = "boeing.boeing.ai/model-provider"
	MCPServerFinalizer             = "boeing.boeing.ai/mcp-server"
	MCPServerCatalogEntryFinalizer = "boeing.boeing.ai/mcp-server-catalog-entry"
	MCPServerInstanceFinalizer     = "boeing.boeing.ai/mcp-server-instance"
	MCPSessionFinalizer            = "boeing.boeing.ai/mcp-session"
	OAuthClientFinalizer           = "boeing.boeing.ai/oauth-client"
	AccessControlRuleFinalizer     = "boeing.boeing.ai/access-control-rule"
	SystemMCPServerFinalizer       = "boeing.boeing.ai/system-mcp-server"
	BoeingbotAgentFinalizer          = "boeing.boeing.ai/boeingbot-agent"
	ImagePullSecretFinalizer       = "boeing.boeing.ai/image-pull-secret"

	ModelProviderSyncAnnotation               = "boeing.ai/model-provider-sync"
	AuthProviderSyncAnnotation                = "boeing.ai/auth-provider-sync"
	MCPCatalogSyncAnnotation                  = "boeing.ai/mcp-catalog-sync"
	SystemMCPCatalogSyncAnnotation            = "boeing.ai/system-mcp-catalog-sync"
	SkillRepositorySyncAnnotation             = "boeing.ai/skill-repository-sync"
	MCPServerCatalogEntrySyncAnnotation       = "boeing.ai/mcp-server-catalog-entry-sync"
	SystemMCPServerCatalogEntrySyncAnnotation = "boeing.ai/system-mcp-server-catalog-entry-sync"
)
