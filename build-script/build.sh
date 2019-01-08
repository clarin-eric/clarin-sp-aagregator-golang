#!/bin/bash

set -e

PRECOMPILED_BUILD_IMAGE="registry.gitlab.com/clarin-eric/build-image:1.1.0"
BUILD_IMAGE=${PRECOMPILED_BUILD_IMAGE}

#
# Set default values for parameters
#
MODE="gitlab"
BUILD=0
TEST=0
RUN=0
RELEASE=0
VERBOSE=0
NO_EXPORT=0
DOWN=0

#
# Process script arguments
#
while [[ $# -gt 0 ]]
do
key="$1"
case $key in
    -b|--build)
        BUILD=1
        ;;
    -d|--down)
        DOWN=1
        ;;
    -h|--help)
        MODE="help"
        ;;
    -l|--local)
        MODE="local"
        ;;
    -n|--no-export)
        NO_EXPORT=1
        ;;
    -r|--release)
        RELEASE=1
        ;;
    -t|--test)
        TEST=1
        ;;
    -r|--run)
        RUN=1
        ;;
    -v|--verbose)
        VERBOSE=1
        ;;
    *)
        echo "Unkown option: $key"
        MODE="help"
        ;;
esac
shift # past argument or value
done

# Source variables if it exists
if [ -f variables.sh ]; then
    if  [ "${NO_EXPORT}" -eq 0 ]; then
        echo "Overriding configuration"
        set -x
    fi
    . ./variables.sh
    set +x
fi

_IMAGE_DIR=${IMAGE_DIR:-"image/"}


# Print parameters if running in verbose mode
if [ ${VERBOSE} -eq 1 ]; then
    echo "build=${BUILD}"
    set -x
fi

source ./copy_data.sh
source ./update_version.sh

#
# Force local mode if run or down commands are specified
#
if [ ${RUN} -eq 1 ] || [ ${DOWN} -eq 1 ]; then
    if [ "${MODE}" != "local" ]; then
        echo "Forcing local mode"
        MODE="local"
    fi
fi

#
# Execute based on mode argument
#
if [ ${MODE} == "help" ]; then
    echo ""
    echo "build.sh [-bdhlnRrtv]"
    echo ""
    echo "  -b, --build      Build docker image"
    echo "  -r, --release    Push docker image to registry"
    echo "  -t, --test       Execute tests"
    echo "  -R, --run        Run (only works in local (-l, --local) mode)"
    echo "  -d, --down       Down (only works in local (-l, --local) mode)"
    echo ""
    echo "  -l, --local      Run workflow locally in a local docker container"
    echo "  -v, --verbose    Run in verbose mode"
    echo "  -n, --no-export  Don't export the build artiface, this is used when running"
    echo "                   the build workflow locally"
    echo ""
    echo "  -h, --help       Show help"
    echo ""
    exit 0
elif [ "${MODE}" == "gitlab" ]; then

    if [ -n "$CI_SERVER" ]; then
        TAG="${CI_BUILD_TAG:-$CI_BUILD_REF}"
        IMAGE_QUALIFIED_NAME="${CI_REGISTRY_IMAGE}:${TAG}"
        IMAGE_FILE_NAME="${CI_REGISTRY_IMAGE##*/}:${TAG}"
    else
        # WARNING: The current working dir must equal the project root dir.

        PROJECT_NAME="$(basename "$(pwd)")"
        TAG="$(git describe --always)"
        IMAGE_QUALIFIED_NAME="$PROJECT_NAME:${TAG:-latest}"
        IMAGE_FILE_NAME="${IMAGE_QUALIFIED_NAME}"
    fi

    #Create output directory on-the-fly
    if [ ! -d './output' ]; then
        mkdir -p 'output'
    fi

    IMAGE_FILE_PATH="$(readlink -fn './output/')/$IMAGE_FILE_NAME.tar.gz"
    export IMAGE_QUALIFIED_NAME
    export IMAGE_FILE_PATH

    #Build
    if [ "${BUILD}" -eq 1 ]; then

        echo "**** Building image ****"
        cd -- ${_IMAGE_DIR}
        if  [ "${NO_EXPORT}" -eq 0 ]; then
            init_data
        fi

        #Build docker build arguments
        DOCKER_ARGS="--tag=$IMAGE_QUALIFIED_NAME"
        if [[ ! -z ${DIST_VERSION+x} ]]; then
            DOCKER_ARGS=${DOCKER_ARGS}" --build-arg DIST_VERSION=${DIST_VERSION}"
        fi

        update_version_before ${TAG}
        docker build ${DOCKER_ARGS} .
        update_version_before ${TAG}

        if  [ "${NO_EXPORT}" -eq 0 ]; then
            cleanup_data
            #Only export artifact to disk when running in a gitlab CI pipeline
            docker save --output="$IMAGE_FILE_PATH" "$IMAGE_QUALIFIED_NAME"
        fi
    fi

     #Test
    if [ "${TEST}" -eq 1 ]; then
        echo "**** Testing image *******************************"
        if [ ! -d 'test' ]; then
            echo "Test directory (./test/) not found"
            exit 1
        fi
        if [ ! -f 'test/docker-compose.yml' ]; then
            echo "docker-compose.yml not found in test directory (./test/docker-compose.yml)"
            exit 1
        fi
        cd -- 'test/'
        apk --quiet update --update-cache
        apk --quiet add 'py2-pip=9.0.0-r1'
        pip --quiet --disable-pip-version-check install 'docker-compose==1.8.0'
        #Load image in gitlab CI pipeline
        if  [ "${NO_EXPORT}" -eq 0 ]; then
            docker load --input="$IMAGE_FILE_PATH"
        fi
        #cleanup to ensure clean state
        docker-compose down -v
        #Start services
        docker-compose up
        #Verify all containers are closed nicely
        number_of_failed_containers="$(docker-compose ps -q | xargs docker inspect \
            -f '{{ .State.ExitCode }}' | grep -c 0 -v | tr -d ' ')"
        #cleanup
        docker-compose down -v
        #return result
        exit "$number_of_failed_containers"
    fi

    if [ "${RUN}" -eq 1 ]; then
        echo "Run is not supported in gitlab mode"
    fi

    #Release
    if [ "${RELEASE}" -eq 1 ]; then
        echo "**** Releasing image ****"
        docker login -u 'gitlab-ci-token' -p "${CI_BUILD_TOKEN}" 'registry.gitlab.com'
        docker load --input="${IMAGE_FILE_PATH}"
        docker push "${IMAGE_QUALIFIED_NAME}"
        docker logout 'registry.gitlab.com'
    fi

elif [ "${MODE}" == "local" ]; then

    if [ "${DOWN}" -eq 1 ]; then
        if [ ! -d 'run' ]; then
            echo "Run directory (./run/) not found"
            exit 1
        fi
        if [ ! -f 'run/docker-compose.yml' ]; then
            echo "docker-compose.yml not found in run directory (./run/docker-compose.yml)"
            exit 1
        fi
        PROJECT_NAME="$(basename "$(pwd)")"
        TAG="$(git describe --always)"
        IMAGE_QUALIFIED_NAME="$PROJECT_NAME:${TAG:-latest}"
        export IMAGE_QUALIFIED_NAME

        cd -- 'run/'
        #cleanup to ensure clean state
        docker-compose down -v
    fi

    #Run
    if [ "${RUN}" -eq 1 ]; then
        if [ ! -d 'run' ]; then
            echo "Run directory (./run/) not found"
            exit 1
        fi
        if [ ! -f 'run/docker-compose.yml' ]; then
            echo "docker-compose.yml not found in run directory (./run/docker-compose.yml)"
            exit 1
        fi
        PROJECT_NAME="$(basename "$(pwd)")"
        TAG="$(git describe --always)"
        IMAGE_QUALIFIED_NAME="$PROJECT_NAME:${TAG:-latest}"
        export IMAGE_QUALIFIED_NAME

        cd -- 'run/'
        #cleanup to ensure clean state
        docker-compose down -v
        #rebuild docker images
        docker-compose build
        #Start services
        docker-compose up
    else
        #
        # Setup all commands
        #

        SHELL_FLAGS=""
        if [ ${VERBOSE} -eq 1 ]; then
            FLAGS="-x"
        fi

        FLAGS=""

        CMD=""
        if [ ${BUILD} -eq 1 ] && [ ${TEST} -eq 1 ]; then
            CMD="sh ${SHELL_FLAGS} ./build.sh --build --no-export ${FLAGS} && sh ${SHELL_FLAGS} ./build.sh --test --no-export ${FLAGS}"
        elif [ ${BUILD} -eq 1 ]; then
            CMD="sh ${SHELL_FLAGS} ./build.sh --build --no-export ${FLAGS}"
        elif [ ${TEST} -eq 1 ]; then
            CMD="sh ${SHELL_FLAGS} ./build.sh --test --no-export ${FLAGS}"
        fi

        #
        # Build process
        #

        #Prepare environmt by downloading external resources when building an image
        if [ ${BUILD} -eq 1 ]; then
            cwd=$(pwd)
            cd -- ${_IMAGE_DIR}
            cleanup_data
            init_data "local"
            cd -- "${cwd}"
        fi

        #Start the build process
        docker run \
            --volume='/var/run/docker.sock:/var/run/docker.sock' \
            --rm \
            --volume="$PWD":"$PWD" \
            --workdir="$PWD" \
            -it \
            ${BUILD_IMAGE} \
            sh -c "${CMD}"

        #Cleanup environment from downloaded resources when building an image
        if [ ${BUILD} -eq 1 ]; then
            cwd=$(pwd)
            cd -- "${_IMAGE_DIR}"
            cleanup_data
            cd -- "${cwd}"
        fi
    fi
else
    exit 1
fi
