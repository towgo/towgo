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
    *)
    echo '目标 ' $3 ' 无法编译'
    exit
    ;;
esac

buildfile=$1

 
case $2 in
    "linux")  echo '交叉编译目标为linux+'$goarch
    echo "CGO_ENABLED=0 GOOS=linux GOARCH="$goarch" go build "$1
    CGO_ENABLED=0 GOOS=linux GOARCH=$goarch go build -o ./"${buildfile%%.*}-"$releasePath $1
    break      
    ;;
    "windows")  echo '交叉编译目标为windows+'$goarch
    echo "CGO_ENABLED=0 GOOS=windows GOARCH="$goarch" go build "$1
    CGO_ENABLED=0 GOOS=windows GOARCH=$goarch go build  -o ././"${buildfile%%.*}-"$releasePath".exe"  $1
    break
    ;;
    "darwin")  echo '交叉编译目标为maxos+'$goarch
    echo "CGO_ENABLED=0 GOOS=darwin GOARCH="$goarch" go build "$1
    CGO_ENABLED=0 GOOS=darwin GOARCH=$goarch go build -o ./"${buildfile%%.*}-"$releasePath $1
    break
    ;;
    *)  echo '参数不正确:'$goarch
    break
    ;;
esac
 echo '脚本执行结束...'
