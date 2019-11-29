# Install qemu: https://wiki.qemu.org/Hosts/Linux
KERNEL=kernel-qemu-4.19.50-buster       # From https://github.com/dhruvvyas90/qemu-rpi-kernel/raw/master/kernel-qemu-4.19.50-buster
DTB=versatile-pb.dtb                    # From https://github.com/dhruvvyas90/qemu-rpi-kernel/raw/master/versatile-pb.dtb
IMG=2019-09-26-raspbian-buster-lite.img # From https://downloads.raspberrypi.org/raspbian_lite_latest

qemu-system-arm \
  -M versatilepb \
  -cpu arm1176 \
  -m 256 \
  -net nic \
  -net user,hostfwd=tcp::5022-:22 \
  -hda $IMG \
  -dtb $DTB \
  -kernel $KERNEL \
  -append 'root=/dev/sda2 panic=1' \
  -display none \
  -serial mon:stdio \
  -no-reboot
