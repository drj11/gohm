for a in $(find . -name '*.go' -exec dirname {} \; | sort)
do (
  cd $a
  go fmt &&
  go install
) done
