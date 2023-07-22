#! /bin/bash

DIR="/home/ec2-user"
MIN_SIZE="10KB"
DESKTOP_QUALITY=90
MOBILE_QUALITY=50
WORKERS=8

# make the desktop webps
/usr/bin/gowebp -d ${DIR} -q ${DESKTOP_QUALITY} -w ${WORKERS} -min-size ${MIN_SIZE} > /var/log/gowebp_desktop.log
# make the mobile webps
/usr/bin/gowebp -d ${DIR} -q ${MOBILE_QUALITY} -w ${WORKERS} -append ".mobile" -min-size ${MIN_SIZE} > /var/log/gowebp_mobile.log

