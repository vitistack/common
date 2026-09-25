{{/*
Expand the name of the chart.
*/}}
{{- define "vitistack-crds.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "vitistack-crds.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "vitistack-crds.labels" -}}
helm.sh/chart: {{ include "vitistack-crds.chart" . }}
{{ include "vitistack-crds.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "vitistack-crds.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vitistack-crds.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "vitistack-crds.annotations" -}}
{{- if .Values.crds.keep }}
helm.sh/resource-policy: keep
{{- end }}
{{- with .Values.annotations }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Fully qualified app name.
*/}}
{{- define "vitistack-crds.fullname" -}}
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
Conversion webhook resource name. Shared by the Deployment, Service,
ServiceAccount, Certificate and the CRD's conversion clientConfig so they
never drift apart.
*/}}
{{- define "vitistack-crds.webhook.name" -}}
{{- printf "%s-conversion-webhook" (include "vitistack-crds.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Conversion webhook selector labels (immutable — used by the Deployment and
Service selectors, must match the pod template labels exactly).
*/}}
{{- define "vitistack-crds.webhook.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vitistack-crds.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: conversion-webhook
{{- end }}

{{/*
Conversion webhook labels.
*/}}
{{- define "vitistack-crds.webhook.labels" -}}
{{ include "vitistack-crds.labels" . }}
app.kubernetes.io/component: conversion-webhook
{{- end }}
