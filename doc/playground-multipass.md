# Playground Setup using `multipass`

Launch a [multipass](https://multipass.run/) LTS instance with this cloud-init script. It installs the recent version of ipvsctl and clones the repository as well.

```bash
$ cat >multipass-cloudinit.yaml <<EOF
bootcmd:
    - apt-get update -y && apt-get install -y -q ipvsadm bats
    - systemctl stop snapd multipathd unattended-upgrades
    - wget https://github.com/aschmidt75/ipvsctl/releases/download/v0.2.2/ipvsctl_0.2.2_Linux_x86_64.tar.gz -O ipvsctl.tar.gz
    - tar -xzf ipvsctl.tar.gz && chmod +x ipvsctl && mv ipvsctl /usr/local/bin && ipvsctl --version
    - git clone https://github.com/aschmidt75/ipvsctl.git
EOF
$ multipass launch -c 1 -m 512M -n ipvsctlplayground --cloud-init multipass-cloudinit.yaml lts;
```

Testing usign the bats-based test suite:

```bash
$ multipass shell ipvsctlplayground
$ sudo -i
# cd ipvsctl/tests
# bats .
(...)
```

