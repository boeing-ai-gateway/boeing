{{/*
Return the chart name and version.
*/}}
{{- define "boeing.chart" -}}
{{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
{{- end -}}

{{/*
Expand the name of the chart.
*/}}
{{- define "boeing.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a fullname using the release name and the chart name.
*/}}
{{- define "boeing.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "boeing.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Create labels for the resources.
*/}}
{{- define "boeing.labels" -}}
helm.sh/chart: {{ include "boeing.chart" . }}
{{ include "boeing.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/component: gateway
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: boeing
{{- if .Values.additionalLabels }}
{{ toYaml .Values.additionalLabels }}
{{- end }}
{{- end -}}

{{/*
Create selector labels for the resources.
*/}}
{{- define "boeing.selectorLabels" -}}
app.kubernetes.io/name: {{ include "boeing.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "boeing.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "boeing.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Set name of secret to use for credentials
*/}}
{{- define "boeing.config.secretName" -}}
{{- if .Values.config.existingSecret -}}
{{- .Values.config.existingSecret -}}
{{- else -}}
{{ .Release.Name }}-config
{{- end -}}
{{- end -}}

{{/*
Set name of namespace to use for mcp servers
*/}}
{{- define "boeing.config.mcpNamespace" -}}
{{- if .Values.mcpNamespace.name -}}
{{- .Values.mcpNamespace.name -}}
{{- else -}}
{{ .Release.Name }}-mcp
{{- end -}}
{{- end -}}

{{/*
Generate comma-separated list of MCP image pull secret names
*/}}
{{- define "boeing.config.mcpImagePullSecrets" -}}
{{- $secrets := list -}}
{{- range .Values.mcpImagePullSecrets -}}
{{- $secrets = append $secrets .name -}}
{{- end -}}
{{- join "," $secrets -}}
{{- end -}}

{{/*
Validate PSA level value. Valid values are: privileged, baseline, restricted
Usage: {{ include "boeing.validatePSALevel" (dict "value" .Values.mcpNamespace.podSecurity.enforce "field" "mcpNamespace.podSecurity.enforce") }}
*/}}
{{- define "boeing.validatePSALevel" -}}
{{- $validLevels := list "privileged" "baseline" "restricted" -}}
{{- if not (has .value $validLevels) -}}
{{- fail (printf "Invalid PSA level %q for %s: must be one of [privileged, baseline, restricted]" .value .field) -}}
{{- end -}}
{{- end -}}

{{/*
Validate all PSA level values in mcpNamespace.podSecurity
*/}}
{{- define "boeing.validatePodSecurityLevels" -}}
{{- if .Values.mcpNamespace.podSecurity.enabled -}}
{{- include "boeing.validatePSALevel" (dict "value" .Values.mcpNamespace.podSecurity.enforce "field" "mcpNamespace.podSecurity.enforce") -}}
{{- include "boeing.validatePSALevel" (dict "value" .Values.mcpNamespace.podSecurity.audit "field" "mcpNamespace.podSecurity.audit") -}}
{{- include "boeing.validatePSALevel" (dict "value" .Values.mcpNamespace.podSecurity.warn "field" "mcpNamespace.podSecurity.warn") -}}
{{- end -}}
{{- end -}}

{{/*
Validate network policy provider Helm chart configuration.
*/}}
{{- define "boeing.validateNetworkPolicyProviderChartConfig" -}}
{{- $repo := .Values.config.BOEING_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_REPO | default "" | toString | trim -}}
{{- $name := .Values.config.BOEING_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_NAME | default "" | toString | trim -}}
{{- if and $repo (not $name) -}}
{{- fail "config.BOEING_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_NAME is required when config.BOEING_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_REPO is set" -}}
{{- end -}}
{{- if and $name (not $repo) -}}
{{- fail "config.BOEING_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_REPO is required when config.BOEING_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_NAME is set" -}}
{{- end -}}
{{- end -}}

{{/*
Get the image tag, defaulting to appVersion. If appVersion looks like a development version (0.0.0-dev),
defaults to "main" tag instead.
*/}}
{{- define "boeing.imageTag" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- if and (not .Values.image.tag) (hasPrefix "0.0.0" .Chart.AppVersion) -}}
{{- $tag = "main" -}}
{{- end -}}
{{- $tag -}}
{{- end -}}

{{/*
Get the MCP base image with tag. If the configured image doesn't contain a tag (no colon after the last slash),
appends the chart's appVersion as the tag. If appVersion looks like a development version (0.0.0-dev),
defaults to "main" tag instead.
*/}}
{{- define "boeing.config.mcpBaseImage" -}}
{{- $image := .Values.config.BOEING_SERVER_MCPBASE_IMAGE -}}
{{- if $image -}}
{{- $parts := splitList "/" $image -}}
{{- $lastPart := last $parts -}}
{{- if contains ":" $lastPart -}}
{{- $image -}}
{{- else -}}
{{- $tag := .Chart.AppVersion -}}
{{- if hasPrefix "0.0.0" $tag -}}
{{- $tag = "main" -}}
{{- end -}}
{{- printf "%s:%s" $image $tag -}}
{{- end -}}
{{- end -}}
{{- end -}}
