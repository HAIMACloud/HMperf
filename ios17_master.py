################################################################################
#
# Copyright (c) 2022 haimacloud.com, Inc. All Rights Reserved
#
################################################################################
"""
This module provide Upload of test results

Authors:chenlei
Date:2024-08-08
"""
import time
import logging

from pymobiledevice3.remote.common import TunnelProtocol
from pymobiledevice3.tunneld import TUNNELD_DEFAULT_ADDRESS, TunneldRunner
from functools import partial

logger = logging.getLogger(__name__)
TUNNELD_DEFAULT_ADDRESS = ('127.0.0.1', 28100)

def startTunnel():
    protocol = TunnelProtocol(TunnelProtocol.QUIC)
    host = TUNNELD_DEFAULT_ADDRESS[0]
    port = TUNNELD_DEFAULT_ADDRESS[1]
    usb = True
    wifi = True
    usbmux = True
    tunneld_runner = partial(TunneldRunner.create, host, port, protocol=protocol, usb_monitor=usb, wifi_monitor=wifi,
                             usbmux_monitor=usbmux)
    tunneld_runner()
    pass
if __name__ == "__main__":
    startTunnel()
    pass