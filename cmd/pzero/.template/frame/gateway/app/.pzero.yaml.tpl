{{ if ne .Style "gozero" }}style: {{.Style}}

{{ end }}gen:
    hooks:
        after:
            - pzero gen swagger