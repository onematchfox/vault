pid_file = "./pidfile"

cache {
    persist = {
        exit_on_err = false
        remove_after_import = false
        path = "/tmp/bolt-file.db"
    }
}

listener "tcp" {
    address = "127.0.0.1:8300"
    tls_disable = true
}
