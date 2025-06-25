{{/* Generate chart name */}}
{{- define "maxmind-geoip-authz.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Generate full name */}}
{{- define "maxmind-geoip-authz.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" (include "maxmind-geoip-authz.name" .) .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/* Chart name and version */}}
{{- define "maxmind-geoip-authz.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
{{- end -}}

{{/* Common labels */}}
{{- define "maxmind-geoip-authz.labels" -}}
helm.sh/chart: {{ include "maxmind-geoip-authz.chart" . }}
{{ include "maxmind-geoip-authz.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/* Selector labels */}}
{{- define "maxmind-geoip-authz.selectorLabels" -}}
app.kubernetes.io/name: {{ include "maxmind-geoip-authz.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* Service account name */}}
{{- define "maxmind-geoip-authz.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- if .Values.serviceAccount.name -}}
{{- .Values.serviceAccount.name -}}
{{- else -}}
{{ include "maxmind-geoip-authz.fullname" . }}
{{- end -}}
{{- else -}}
{{- if .Values.serviceAccount.name -}}
{{- .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}
{{- end -}}
