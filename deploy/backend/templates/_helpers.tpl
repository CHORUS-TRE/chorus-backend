{{- define "backend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "backend.nameid" -}}
{{- printf "%s-%s" (include "backend.name" .) .Values.version }}
{{- end }}

{{- define "backend.main" -}}
    {{- $passphrase := "" }}
    {{- if and (.Values.aesPassphrase) (ne .Values.aesPassphrase "") }}
        {{- $passphrase = .Values.aesPassphrase }}
    {{- else }}
        {{- $namespace := "backend" }}
        {{- $secretName := "configaespassphrase-argodev" }}
        {{- $secretKey := "passphrase" }}
        {{- $secret := lookup "v1" "Secret" $namespace $secretName }}
        {{- if $secret }}
            {{- $passphrase = $secret.data.passphrase | b64dec }}
        {{- end }}
    {{- end }}

    {{- if (ne $passphrase "") }}
        {{- $encryptedFile := .Files.Get "files/secrets.yaml.enc" }}
        {{- $salt := $encryptedFile | substr 0 64 }}
        {{- $remainingFile := $encryptedFile | substr 64 -1 }}
        {{- $ivAndEncryptedData := $remainingFile | b64enc }}
        {{- $passphraseWithSalt :=  (printf "%s%s" $passphrase $salt) | sha256sum | substr 0 32 }}
        {{- $secrets := $ivAndEncryptedData | decryptAES $passphraseWithSalt | fromYaml -}}

        {{- $_ := set $secrets "Template" $.Template }}
        {{- tpl (.Files.Get "files/main.yaml") $secrets -}}
    {{- else }}
        {{- $file := .Files.Get "files/main.yaml" }}
        {{- if $file }}
            {{- printf "%s" (.Files.Get "files/main.yaml") }}
        {{- else }}
            {{- printf "%s" ((.Values.main).yaml) }}
        {{- end }}
    {{- end }}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "backend.fullname" -}}
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
{{- define "backend.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "backend.labels" -}}
helm.sh/chart: {{ include "backend.chart" . }}
{{ include "backend.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "backend.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "backend.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
