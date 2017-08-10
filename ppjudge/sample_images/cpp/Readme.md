# CPP Image for ppjudge template

## lang.json template
```
{"compile_image":"ppjudge_cpp:gcc5.4.0","exec_image":"ppjudge_cpp:gcc5.4.0","compile_command":["g++","-std=c++14","-O2","-o","/work/a.out","/work/main.cpp"],"exec_command":["/work/a.out"],"source_file_name":"main.cpp", "compile_env":[], "exec_env":[]}
```
## Installation
- Edit Dockerfile and change image tag
- ./build.sh
