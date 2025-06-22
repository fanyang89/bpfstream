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
```
