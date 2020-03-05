# cleaner

특정 디렉토리의 디스크 여유공간을 유지하기 위해 오래된 파일순으로 삭제

- 로컬 디스크일 경우 file notification 활용
- free 용량 체크를 위해 파티션 정보를 읽어야 함
- directory 와 파일 생성/삭제만 검사

## fsnotify 주의점

mkdir -p로 만든 하위폴더는 감지 안됨

## Benchmark scan speed

4가지 라이브러리 성능 테스트

- filepath.Walk
- ioutil.Readdir
- os.File.Readdir
- godirwalk

FileInfo 를 읽는 것에 속도 영향이 큼. godirwalk 는 FileInfo 읽지 않음

```sh
# MacOS 에서 테스트
go test -bench=. -benchmem

Benchmark_GoDirWalk-8                  2         809993550 ns/op        46439280 B/op     208014 allocs/op
Benchmark_FilePathWalkDir1-8           2         578208854 ns/op        45451488 B/op     216355 allocs/op
Benchmark_IOReadDir1-8             26961             44624 ns/op            4192 B/op         26 allocs/op
Benchmark_OSReadDir1-8             27537             43440 ns/op            4096 B/op         23 allocs/o
```
