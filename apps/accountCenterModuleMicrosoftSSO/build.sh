case $3 in
    "")
    goarch="amd64"
    releasePath="x86_64"
    break
    ;;
    "arm")
    goarch="arm"
    releasePath="arm"
    break
    ;;
    "arm64")
    goarch="arm64"
    releasePath="arm64"
    break
    ;;
    "loong64")
    goarch="loong64"
    releasePath="loong64"
    break
    ;;
    "up")
        goarch="amd64"
        releasePath="x86_64"
        isUpload=true
      break
      ;;
    *)
    echo '目标 ' $3 ' 无法编译'
    exit
    ;;
esac

buildfile=$1


case $2 in
    "l"|"linux"|"")  echo '交叉编译目标为linux+'$goarch
    echo "CGO_ENABLED=0 GOOS=linux GOARCH="$goarch" go build "$1
    CGO_ENABLED=0 GOOS=linux GOARCH=$goarch go build -o ./"${buildfile%.*}-"$releasePath $1
    break
    ;;
    "w"|"windows")  echo '交叉编译目标为windows+'$goarch
    echo "CGO_ENABLED=0 GOOS=windows GOARCH="$goarch" go build "$1
    CGO_ENABLED=0 GOOS=windows GOARCH=$goarch go build  -o ././"${buildfile%.*}-"$releasePath".exe"  $1
    break
    ;;
    "d"|"m"|"darwin")  echo '交叉编译目标为maxos+'$goarch
    echo "CGO_ENABLED=0 GOOS=darwin GOARCH="$goarch" go build "$1
    CGO_ENABLED=0 GOOS=darwin GOARCH=$goarch go build -o ./"${buildfile%.*}-"$releasePath $1
    break
    ;;
    "up")
      echo '交叉编译目标为linux+'$goarch
      isUpload=true
      echo "CGO_ENABLED=0 GOOS=linux GOARCH="$goarch" go build "$1
      CGO_ENABLED=0 GOOS=linux GOARCH=$goarch go build -o ./"${buildfile%.*}-"$releasePath $1
      break
      ;;
    *)  echo '参数不正确:'$goarch
    break
    ;;
esac

case $4 in
  "up")
  isUpload=true
  ;;
esac

if [ $isUpload ]; then
     echo '正在将文件'"${buildfile%.*}-"$releasePath'上传至 172.0.0.19 ...'
     ssh root@172.0.0.19 "rm -rf /home/workspace/ysld_digital_employee/ysld.fanhaninfo.test/accountCenter/start-x86_64"
     wait
     mv "${buildfile%.*}-$releasePath" start-x86_64
     wait
     scp -C start-x86_64 root@172.0.0.19:/home/workspace/ysld_digital_employee/ysld.fanhaninfo.test/accountCenter/ &
     wait
     echo '文件上传完毕'
fi
 echo '脚本执行结束...'
