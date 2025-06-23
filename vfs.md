```
kretfunc:vmlinux:vfs_read
    struct file * file
    char * buf
    size_t count
    loff_t * pos
    ssize_t retval
kretfunc:vmlinux:vfs_readlink
    struct dentry * dentry
    char * buffer
    int buflen
    int retval
kretfunc:vmlinux:vfs_readv
    struct file * file
    const struct iovec * vec
    long unsigned int vlen
    loff_t * pos
    rwf_t flags
    ssize_t retval
kretfunc:vmlinux:vfs_write
    struct file * file
    const char * buf
    size_t count
    loff_t * pos
    ssize_t retval
kretfunc:vmlinux:vfs_writev
    struct file * file
    const struct iovec * vec
    long unsigned int vlen
    loff_t * pos
    rwf_t flags
    ssize_t retval
kretfunc:vmlinux:vfs_fsync
    struct file * file
    int datasync
    int retval
kretfunc:vmlinux:vfs_fsync_range
    struct file * file
    loff_t start
    loff_t end
    int datasync
    int retval
kretfunc:vmlinux:vfs_open
    const struct path * path
    struct file * file
    int retval
kretfunc:vmlinux:vfs_create
    struct mnt_idmap * idmap
    struct inode * dir
    struct dentry * dentry
    umode_t mode
    bool want_excl
    int retval

-> % sudo bpftrace -lv 'kretfunc:dentry_open'
fexit:vmlinux:dentry_open
    const struct path * path
    int flags
    const struct cred * cred
    struct file * retval
```

```
-> % sudo bpftrace -f json vfs-raw.bt
{"type": "attached_probes", "data": {"probes": 6}}
vfs-raw.bt:10:80-96: ERROR: helper bpf_d_path not allowed in probe
  printf("ts=%lld, fn=vfs_open, rc=%d, pid=%d, path='%s'", nsecs, retval, pid, path(args->path));
```

```
fentry:vmlinux:vfs_getattr
    const struct path * path
    struct kstat * stat
    u32 request_mask
    unsigned int query_flags
    int retval
```

```
BTF_SET_START(btf_allowlist_d_path)
#ifdef CONFIG_SECURITY
BTF_ID(func, security_file_permission)
BTF_ID(func, security_inode_getattr)
BTF_ID(func, security_file_open)
#endif
#ifdef CONFIG_SECURITY_PATH
BTF_ID(func, security_path_truncate)
#endif
BTF_ID(func, vfs_truncate)
BTF_ID(func, vfs_fallocate)
BTF_ID(func, dentry_open)
BTF_ID(func, vfs_getattr)
BTF_ID(func, filp_close)
BTF_SET_END(btf_allowlist_d_path)
```

```
{"type": "lost_events", "data": {"events": 67190}}
```
