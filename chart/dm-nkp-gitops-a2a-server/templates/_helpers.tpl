{{/*
Expand the name of the chart.
*/}}
{{- define "dm-nkp-gitops-a2a-server.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "dm-nkp-gitops-a2a-server.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "dm-nkp-gitops-a2a-server.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "dm-nkp-gitops-a2a-server.labels" -}}
helm.sh/chart: {{ include "dm-nkp-gitops-a2a-server.chart" . }}
{{ include "dm-nkp-gitops-a2a-server.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "dm-nkp-gitops-a2a-server.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dm-nkp-gitops-a2a-server.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "dm-nkp-gitops-a2a-server.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "dm-nkp-gitops-a2a-server.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
TLS secret name
*/}}
{{- define "dm-nkp-gitops-a2a-server.tlsSecretName" -}}
{{- if .Values.tls.secretName }}
{{- .Values.tls.secretName }}
{{- else }}
{{- printf "%s-tls" (include "dm-nkp-gitops-a2a-server.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Base URL for agent card
*/}}
{{- define "dm-nkp-gitops-a2a-server.baseURL" -}}
{{- if .Values.a2a.baseURL }}
{{- .Values.a2a.baseURL }}
{{- else if .Values.tls.enabled }}
{{- printf "https://%s" .Values.httpRoute.hostname }}
{{- else }}
{{- printf "http://%s" .Values.httpRoute.hostname }}
{{- end }}
{{- end }}
