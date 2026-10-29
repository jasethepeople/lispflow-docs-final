{{- define "lispflow.name" -}}lispflow{{- end -}}
{{- define "lispflow.fullname" -}}lispflow{{- end -}}
{{- define "lispflow.labels" -}}
app.kubernetes.io/name: {{ include "lispflow.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
