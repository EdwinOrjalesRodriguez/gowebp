#! /bin/bash

DIR="/some/dir/containing/images"
MIN_SIZE="10KB"
DESKTOP_QUALITY=90
MOBILE_QUALITY=50
WORKERS=8
CHOWN="webapp"
CHGRP="webapp"
CHMOD=0775

# make the desktop webps
/usr/bin/gowebp -d ${DIR} -q ${DESKTOP_QUALITY} -w ${WORKERS} -min-size ${MIN_SIZE} -chown ${CHOWN} -chgrp ${CHGRP} -chmod ${CHMOD}
# make the mobile webps
/usr/bin/gowebp -d ${DIR} -q ${MOBILE_QUALITY} -w ${WORKERS} -append ".mobile" -min-size ${MIN_SIZE} -chown ${CHOWN} -chgrp ${CHGRP} -chmod ${CHMOD}