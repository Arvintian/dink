package template

import (
	"encoding/json"
	"fmt"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"

	"github.com/docker/docker/api/types"
	spec "github.com/opencontainers/runtime-spec/specs-go"
)

func GetHostName(container *dinkv1beta1.Container) string {
	return container.Status.ContainerID[:12]
}

func CreateRuntimeConfig(container *dinkv1beta1.Container, image types.ImageInspect) spec.Spec {
	var config spec.Spec
	if err := json.Unmarshal([]byte(runtimeSpec), &config); err != nil {
		panic(err)
	}
	config.Process.Env = append(config.Process.Env, image.Config.Env...)
	config.Process.Args = append(config.Process.Args, image.Config.Entrypoint...)
	config.Process.Args = append(config.Process.Args, image.Config.Cmd...)
	if image.Config.WorkingDir != "" {
		config.Process.Cwd = image.Config.WorkingDir
	}
	if container.Spec.Template.SecurityContext != nil {
		if container.Spec.Template.SecurityContext.RunAsUser != nil {
			config.Process.User.UID = uint32(*container.Spec.Template.SecurityContext.RunAsUser)
		}
		if container.Spec.Template.SecurityContext.RunAsGroup != nil {
			config.Process.User.GID = uint32(*container.Spec.Template.SecurityContext.RunAsGroup)
		}
	}

	config.Hostname = GetPodName(container)
	if container.Spec.HostName != "" {
		config.Hostname = container.Spec.HostName
	}
	envs := []string{}
	for _, item := range container.Spec.Template.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", item.Name, item.Value))
	}
	config.Process.Env = append(config.Process.Env, envs...)
	if container.Spec.Template.WorkingDir != "" {
		config.Process.Cwd = container.Spec.Template.WorkingDir
	}
	if len(container.Spec.Template.Command) > 0 {
		config.Process.Args = container.Spec.Template.Command
		config.Process.Args = append(config.Process.Args, container.Spec.Template.Args...)
	}
	config.Process.Terminal = container.Spec.Template.TTY
	for _, vol := range container.Spec.Template.VolumeMounts {
		config.Mounts = append(config.Mounts, spec.Mount{
			Destination: vol.MountPath,
			Source:      vol.MountPath,
			Type:        "bind",
			Options:     []string{"rbind", "rprivate"},
		})
	}

	return config
}

var runtimeSpec = `
{
    "ociVersion": "1.0.2-dev",
    "process": {
        "user": {
            "uid": 0,
            "gid": 0,
			"additionalGids": []
        },
        "args": [],
        "env": [],
        "cwd": "/",
        "capabilities": {
            "bounding": [
                "CAP_CHOWN",
                "CAP_DAC_OVERRIDE",
                "CAP_FSETID",
                "CAP_FOWNER",
                "CAP_MKNOD",
                "CAP_NET_RAW",
                "CAP_SETGID",
                "CAP_SETUID",
                "CAP_SETFCAP",
                "CAP_SETPCAP",
                "CAP_NET_BIND_SERVICE",
                "CAP_SYS_CHROOT",
                "CAP_KILL",
                "CAP_AUDIT_WRITE"
            ],
            "effective": [
                "CAP_CHOWN",
                "CAP_DAC_OVERRIDE",
                "CAP_FSETID",
                "CAP_FOWNER",
                "CAP_MKNOD",
                "CAP_NET_RAW",
                "CAP_SETGID",
                "CAP_SETUID",
                "CAP_SETFCAP",
                "CAP_SETPCAP",
                "CAP_NET_BIND_SERVICE",
                "CAP_SYS_CHROOT",
                "CAP_KILL",
                "CAP_AUDIT_WRITE"
            ],
            "permitted": [
                "CAP_CHOWN",
                "CAP_DAC_OVERRIDE",
                "CAP_FSETID",
                "CAP_FOWNER",
                "CAP_MKNOD",
                "CAP_NET_RAW",
                "CAP_SETGID",
                "CAP_SETUID",
                "CAP_SETFCAP",
                "CAP_SETPCAP",
                "CAP_NET_BIND_SERVICE",
                "CAP_SYS_CHROOT",
                "CAP_KILL",
                "CAP_AUDIT_WRITE"
            ]
        },
        "oomScoreAdj": 0
    },
    "root": {
        "path": "rootfs"
    },
    "hostname": "runc",
    "mounts": [
        {
            "destination": "/proc",
            "type": "proc",
            "source": "proc",
            "options": [
                "nosuid",
                "noexec",
                "nodev"
            ]
        },
        {
            "destination": "/dev",
            "type": "tmpfs",
            "source": "tmpfs",
            "options": [
                "nosuid",
                "strictatime",
                "mode=755",
                "size=65536k"
            ]
        },
        {
            "destination": "/dev/pts",
            "type": "devpts",
            "source": "devpts",
            "options": [
                "nosuid",
                "noexec",
                "newinstance",
                "ptmxmode=0666",
                "mode=0620",
                "gid=5"
            ]
        },
        {
            "destination": "/sys",
            "type": "sysfs",
            "source": "sysfs",
            "options": [
                "nosuid",
                "noexec",
                "nodev",
                "ro"
            ]
        },
        {
            "destination": "/sys/fs/cgroup",
            "type": "cgroup",
            "source": "cgroup",
            "options": [
                "ro",
                "nosuid",
                "noexec",
                "nodev"
            ]
        },
        {
            "destination": "/dev/mqueue",
            "type": "mqueue",
            "source": "mqueue",
            "options": [
                "nosuid",
                "noexec",
                "nodev"
            ]
        },
        {
            "destination": "/dev/shm",
            "type": "tmpfs",
            "source": "shm",
            "options": [
                "nosuid",
                "noexec",
                "nodev",
                "mode=1777",
                "size=67108864"
            ]
        },
        {
            "destination": "/etc/resolv.conf",
            "type": "bind",
            "source": "/etc/resolv.conf",
            "options": [
                "rbind",
                "rprivate"
            ]
        },
        {
            "destination": "/etc/hostname",
            "type": "bind",
            "source": "/etc/hostname",
            "options": [
                "rbind",
                "rprivate"
            ]
        },
        {
            "destination": "/etc/hosts",
            "type": "bind",
            "source": "/etc/hosts",
            "options": [
                "rbind",
                "rprivate"
            ]
        }
    ],
    "linux": {
        "resources": {
            "devices": [
                {
                    "allow": false,
                    "access": "rwm"
                },
                {
                    "allow": true,
                    "type": "c",
                    "major": 1,
                    "minor": 5,
                    "access": "rwm"
                },
                {
                    "allow": true,
                    "type": "c",
                    "major": 1,
                    "minor": 3,
                    "access": "rwm"
                },
                {
                    "allow": true,
                    "type": "c",
                    "major": 1,
                    "minor": 9,
                    "access": "rwm"
                },
                {
                    "allow": true,
                    "type": "c",
                    "major": 1,
                    "minor": 8,
                    "access": "rwm"
                },
                {
                    "allow": true,
                    "type": "c",
                    "major": 5,
                    "minor": 0,
                    "access": "rwm"
                },
                {
                    "allow": true,
                    "type": "c",
                    "major": 5,
                    "minor": 1,
                    "access": "rwm"
                },
                {
                    "allow": false,
                    "type": "c",
                    "major": 10,
                    "minor": 229,
                    "access": "rwm"
                }
            ],
            "memory": {},
            "cpu": {
                "shares": 0
            },
            "blockIO": {
                "weight": 0
            }
        },
        "namespaces": [
            {
                "type": "mount"
            },
            {
                "type": "uts"
            },
            {
                "type": "pid"
            },
            {
                "type": "ipc"
            }
        ],
        "seccomp": {
            "defaultAction": "SCMP_ACT_ERRNO",
            "syscalls": [
                {
                    "names": [
                        "accept",
                        "accept4",
                        "access",
                        "adjtimex",
                        "alarm",
                        "bind",
                        "brk",
                        "capget",
                        "capset",
                        "chdir",
                        "chmod",
                        "chown",
                        "chown32",
                        "clock_adjtime",
                        "clock_adjtime64",
                        "clock_getres",
                        "clock_getres_time64",
                        "clock_gettime",
                        "clock_gettime64",
                        "clock_nanosleep",
                        "clock_nanosleep_time64",
                        "close",
                        "close_range",
                        "connect",
                        "copy_file_range",
                        "creat",
                        "dup",
                        "dup2",
                        "dup3",
                        "epoll_create",
                        "epoll_create1",
                        "epoll_ctl",
                        "epoll_ctl_old",
                        "epoll_pwait",
                        "epoll_pwait2",
                        "epoll_wait",
                        "epoll_wait_old",
                        "eventfd",
                        "eventfd2",
                        "execve",
                        "execveat",
                        "exit",
                        "exit_group",
                        "faccessat",
                        "faccessat2",
                        "fadvise64",
                        "fadvise64_64",
                        "fallocate",
                        "fanotify_mark",
                        "fchdir",
                        "fchmod",
                        "fchmodat",
                        "fchown",
                        "fchown32",
                        "fchownat",
                        "fcntl",
                        "fcntl64",
                        "fdatasync",
                        "fgetxattr",
                        "flistxattr",
                        "flock",
                        "fork",
                        "fremovexattr",
                        "fsetxattr",
                        "fstat",
                        "fstat64",
                        "fstatat64",
                        "fstatfs",
                        "fstatfs64",
                        "fsync",
                        "ftruncate",
                        "ftruncate64",
                        "futex",
                        "futex_time64",
                        "futimesat",
                        "getcpu",
                        "getcwd",
                        "getdents",
                        "getdents64",
                        "getegid",
                        "getegid32",
                        "geteuid",
                        "geteuid32",
                        "getgid",
                        "getgid32",
                        "getgroups",
                        "getgroups32",
                        "getitimer",
                        "getpeername",
                        "getpgid",
                        "getpgrp",
                        "getpid",
                        "getppid",
                        "getpriority",
                        "getrandom",
                        "getresgid",
                        "getresgid32",
                        "getresuid",
                        "getresuid32",
                        "getrlimit",
                        "get_robust_list",
                        "getrusage",
                        "getsid",
                        "getsockname",
                        "getsockopt",
                        "get_thread_area",
                        "gettid",
                        "gettimeofday",
                        "getuid",
                        "getuid32",
                        "getxattr",
                        "inotify_add_watch",
                        "inotify_init",
                        "inotify_init1",
                        "inotify_rm_watch",
                        "io_cancel",
                        "ioctl",
                        "io_destroy",
                        "io_getevents",
                        "io_pgetevents",
                        "io_pgetevents_time64",
                        "ioprio_get",
                        "ioprio_set",
                        "io_setup",
                        "io_submit",
                        "io_uring_enter",
                        "io_uring_register",
                        "io_uring_setup",
                        "ipc",
                        "kill",
						"landlock_add_rule",
                        "landlock_create_ruleset",
                        "landlock_restrict_self",
                        "lchown",
                        "lchown32",
                        "lgetxattr",
                        "link",
                        "linkat",
                        "listen",
                        "listxattr",
                        "llistxattr",
                        "_llseek",
                        "lremovexattr",
                        "lseek",
                        "lsetxattr",
                        "lstat",
                        "lstat64",
                        "madvise",
                        "membarrier",
                        "memfd_create",
						"memfd_secret",
                        "mincore",
                        "mkdir",
                        "mkdirat",
                        "mknod",
                        "mknodat",
                        "mlock",
                        "mlock2",
                        "mlockall",
                        "mmap",
                        "mmap2",
                        "mprotect",
                        "mq_getsetattr",
                        "mq_notify",
                        "mq_open",
                        "mq_timedreceive",
                        "mq_timedreceive_time64",
                        "mq_timedsend",
                        "mq_timedsend_time64",
                        "mq_unlink",
                        "mremap",
                        "msgctl",
                        "msgget",
                        "msgrcv",
                        "msgsnd",
                        "msync",
                        "munlock",
                        "munlockall",
                        "munmap",
                        "nanosleep",
                        "newfstatat",
                        "_newselect",
                        "open",
                        "openat",
                        "openat2",
                        "pause",
                        "pidfd_open",
                        "pidfd_send_signal",
                        "pipe",
                        "pipe2",
                        "poll",
                        "ppoll",
                        "ppoll_time64",
                        "prctl",
                        "pread64",
                        "preadv",
                        "preadv2",
                        "prlimit64",
						"process_mrelease",
                        "pselect6",
                        "pselect6_time64",
                        "pwrite64",
                        "pwritev",
                        "pwritev2",
                        "read",
                        "readahead",
                        "readlink",
                        "readlinkat",
                        "readv",
                        "recv",
                        "recvfrom",
                        "recvmmsg",
                        "recvmmsg_time64",
                        "recvmsg",
                        "remap_file_pages",
                        "removexattr",
                        "rename",
                        "renameat",
                        "renameat2",
                        "restart_syscall",
                        "rmdir",
                        "rseq",
                        "rt_sigaction",
                        "rt_sigpending",
                        "rt_sigprocmask",
                        "rt_sigqueueinfo",
                        "rt_sigreturn",
                        "rt_sigsuspend",
                        "rt_sigtimedwait",
                        "rt_sigtimedwait_time64",
                        "rt_tgsigqueueinfo",
                        "sched_getaffinity",
                        "sched_getattr",
                        "sched_getparam",
                        "sched_get_priority_max",
                        "sched_get_priority_min",
                        "sched_getscheduler",
                        "sched_rr_get_interval",
                        "sched_rr_get_interval_time64",
                        "sched_setaffinity",
                        "sched_setattr",
                        "sched_setparam",
                        "sched_setscheduler",
                        "sched_yield",
                        "seccomp",
                        "select",
                        "semctl",
                        "semget",
                        "semop",
                        "semtimedop",
                        "semtimedop_time64",
                        "send",
                        "sendfile",
                        "sendfile64",
                        "sendmmsg",
                        "sendmsg",
                        "sendto",
                        "setfsgid",
                        "setfsgid32",
                        "setfsuid",
                        "setfsuid32",
                        "setgid",
                        "setgid32",
                        "setgroups",
                        "setgroups32",
                        "setitimer",
                        "setpgid",
                        "setpriority",
                        "setregid",
                        "setregid32",
                        "setresgid",
                        "setresgid32",
                        "setresuid",
                        "setresuid32",
                        "setreuid",
                        "setreuid32",
                        "setrlimit",
                        "set_robust_list",
                        "setsid",
                        "setsockopt",
                        "set_thread_area",
                        "set_tid_address",
                        "setuid",
                        "setuid32",
                        "setxattr",
                        "shmat",
                        "shmctl",
                        "shmdt",
                        "shmget",
                        "shutdown",
                        "sigaltstack",
                        "signalfd",
                        "signalfd4",
                        "sigprocmask",
                        "sigreturn",
                        "socket",
                        "socketcall",
                        "socketpair",
                        "splice",
                        "stat",
                        "stat64",
                        "statfs",
                        "statfs64",
                        "statx",
                        "symlink",
                        "symlinkat",
                        "sync",
                        "sync_file_range",
                        "syncfs",
                        "sysinfo",
                        "tee",
                        "tgkill",
                        "time",
                        "timer_create",
                        "timer_delete",
                        "timer_getoverrun",
                        "timer_gettime",
                        "timer_gettime64",
                        "timer_settime",
                        "timer_settime64",
                        "timerfd_create",
                        "timerfd_gettime",
                        "timerfd_gettime64",
                        "timerfd_settime",
                        "timerfd_settime64",
                        "times",
                        "tkill",
                        "truncate",
                        "truncate64",
                        "ugetrlimit",
                        "umask",
                        "uname",
                        "unlink",
                        "unlinkat",
                        "utime",
                        "utimensat",
                        "utimensat_time64",
                        "utimes",
                        "vfork",
                        "vmsplice",
                        "wait4",
                        "waitid",
                        "waitpid",
                        "write",
                        "writev"
                    ],
                    "action": "SCMP_ACT_ALLOW"
                },
                {
                    "names": [
                        "ptrace"
                    ],
                    "action": "SCMP_ACT_ALLOW"
                },
				{
                    "names": [
                        "socket"
                    ],
                    "action": "SCMP_ACT_ALLOW",
                    "args": [
                        {
                            "index": 0,
                            "value": 40,
                            "op": "SCMP_CMP_NE"
                        }
                    ]
                },
                {
                    "names": [
                        "personality"
                    ],
                    "action": "SCMP_ACT_ALLOW",
                    "args": [
                        {
                            "index": 0,
                            "value": 0,
                            "op": "SCMP_CMP_EQ"
                        }
                    ]
                },
                {
                    "names": [
                        "personality"
                    ],
                    "action": "SCMP_ACT_ALLOW",
                    "args": [
                        {
                            "index": 0,
                            "value": 8,
                            "op": "SCMP_CMP_EQ"
                        }
                    ]
                },
                {
                    "names": [
                        "personality"
                    ],
                    "action": "SCMP_ACT_ALLOW",
                    "args": [
                        {
                            "index": 0,
                            "value": 131072,
                            "op": "SCMP_CMP_EQ"
                        }
                    ]
                },
                {
                    "names": [
                        "personality"
                    ],
                    "action": "SCMP_ACT_ALLOW",
                    "args": [
                        {
                            "index": 0,
                            "value": 131080,
                            "op": "SCMP_CMP_EQ"
                        }
                    ]
                },
                {
                    "names": [
                        "personality"
                    ],
                    "action": "SCMP_ACT_ALLOW",
                    "args": [
                        {
                            "index": 0,
                            "value": 4294967295,
                            "op": "SCMP_CMP_EQ"
                        }
                    ]
                },
                {
                    "names": [
                        "arch_prctl"
                    ],
                    "action": "SCMP_ACT_ALLOW"
                },
                {
                    "names": [
                        "modify_ldt"
                    ],
                    "action": "SCMP_ACT_ALLOW"
                },
                {
                    "names": [
                        "clone"
                    ],
                    "action": "SCMP_ACT_ALLOW",
                    "args": [
                        {
                            "index": 0,
                            "value": 2114060288,
                            "op": "SCMP_CMP_MASKED_EQ"
                        }
                    ]
                },
                {
                    "names": [
                        "clone3"
                    ],
                    "action": "SCMP_ACT_ERRNO",
                    "errnoRet": 38
                },
                {
                    "names": [
                        "chroot"
                    ],
                    "action": "SCMP_ACT_ALLOW"
                }
            ]
        },
        "maskedPaths": [
            "/proc/asound",
            "/proc/acpi",
            "/proc/kcore",
            "/proc/keys",
            "/proc/latency_stats",
            "/proc/timer_list",
            "/proc/timer_stats",
            "/proc/sched_debug",
            "/proc/scsi",
            "/sys/firmware"
        ],
        "readonlyPaths": [
            "/proc/bus",
            "/proc/fs",
            "/proc/irq",
            "/proc/sys",
            "/proc/sysrq-trigger"
        ]
    }
}
`
