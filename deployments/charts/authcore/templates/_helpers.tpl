{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "authcore.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "authcore.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "authcore.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create first time setup user
*/}}
{{- define "authcore.setup.adminPassword" -}}
{{ randAlphaNum 10 | quote }}
{{- end -}}

{{/*
Get redis configuation
*/}}


{{- define "authcore.redis.passwordSecret" -}}
{{- $redisContext := dict "Values" .Values.redis "Release" .Release "Chart" (dict "Name" "redis") -}}
{{- if $.Values.tags }}
  {{- if $.Values.tags.install_redis }}
    {{- include "redis.fullname" $redisContext }}
  {{- end -}}
{{- else if $.Values.redis -}}
  {{- if $.Values.redis.passwordSecret }}
  {{- $.Values.redis.passwordSecret }}
  {{- end -}}
{{- end -}}
{{- end -}}


{{- define "authcore.redis.sentinel_enabled" -}}
{{- $redisContext := dict "Values" .Values.redis "Release" .Release "Chart" (dict "Name" "redis") -}}
{{- if $.Values.tags }}
  {{- if $.Values.tags.install_redis }}
    {{- $redisContext.Values.sentinel.enabled | quote }}
  {{- end -}}
{{- else if $.Values.redis -}}
  {{- if $.Values.redis.sentinel_enabled -}}
    {{- $.Values.redis.sentinel_enabled }}
  {{- end -}}
{{- end -}}
{{- end -}}

{{- define "authcore.redis.sentinel_address" -}}
{{- $redisContext := dict "Values" .Values.redis "Release" .Release "Chart" (dict "Name" "redis") -}}
{{- if $.Values.tags }}
  {{- if $.Values.tags.install_redis }}
    {{- if $redisContext.Values.sentinel.enabled }}
      {{- include "redis.fullname" $redisContext }}:{{ $.Values.redis.sentinel.service.redisPort }}
    {{- end -}}
  {{- end -}}
{{- else if $.Values.redis -}}
  {{- if eq $.Values.redis.sentinel_enabled "true" -}}
    {{- toYaml $.Values.redis.sentinel_address }}
  {{- end -}}
{{- end -}}
{{- end -}}

{{- define "authcore.redis.redis_address" -}}
{{- $redisContext := dict "Values" .Values.redis "Release" .Release "Chart" (dict "Name" "redis") -}}
{{- if $.Values.redis }}
  {{- if $.Values.redis.redis_address }}
    {{- $.Values.redis.redis_address }}
  {{- end -}}
{{- else if $.Values.tags -}}
  {{- if $.Values.tags.install_redis }}
    {{- include "redis.fullname" $redisContext }}-master:{{ $.Values.redis.master.service.port }}
  {{- end -}}
{{- end -}}
{{- end -}}

{{- define "authcore.mysql.databaseURL" -}}
{{- if $.Values.mysql }}
{{- $.Values.mysql.database_url }}
{{- end -}}
{{- end -}}