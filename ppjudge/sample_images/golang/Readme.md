# Golang Image for ppjudge template

## lang.json template
```
{"compile_image":"ppjudge_golang:1.8","exec_image":"ppjudge_golang:1.8","compile_command":["go","build","-o","/work/a.out","/work/main.go"],"exec_command":["/work/a.out"],"source_file_name":"main.go", "compile_env":[], "exec_env":[]}
```
## Installation
- Edit Dockerfile and change image tag
- ./build.sh