query() {
    OP="$1"
    ROUTE="$2"
    ENDMARKER="$3"
    shift 3 || true

    case "${OP}" in
        PUT|POST|DELETE) ;;
        *)
            echo "BAD OP ${OP}"
            return 1;;
    esac

    declare -a ARGS

    ARGS+=(-X "${OP}" "http://localhost:9200/sink${ROUTE}")
    ARGS+=(-H "Content-Type: application/json")

    case "${ENDMARKER}" in
        NO|END|NODATA|NOBODY);;
        *) ARGS+=(-d @/dev/stdin);;
    esac

    curl -o /dev/tty --silent "${ARGS[@]}"
    echo
}

setup() {
    query DELETE '' END

    query PUT '' <<END
{
    "mappings": {
        "properties": {
            "stamp": {
                "type": "date",
                "format":
                "MMM  d HH:mm:ss.SSS||MMM dd HH:mm:ss.SSS||strict_date_optional_time||epoch_millis"
            },
            "blob": {
                "type":"text"
            }
        }
    }
}
END
}

flush() {
    query POST /_flush END
}

cumsum() {
    query POST /_count <<END
{
    "query": {
        "range": {
            "stamp": {"gte": "$1"}
        }
    }
}
END
}

grepfn() {
    query POST /_count <<END
{
    "query": {
        "term": {
            "stamp": "$1"
        }
    }
}
END
}

case "$1" in
    setup) setup ;;
    flush) flush ;;
    cumsum) cumsum "${2:-"1970-01-01T00:00:00.000Z"}" ;;
    grep) grepfn "${2:?}" ;;
    *)
        echo "bad command \"$1\""
        exit 1
    ;;
esac
