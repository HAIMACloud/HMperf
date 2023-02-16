################################################################################
#
# Copyright (c) 2022 haimacloud.com, Inc. All Rights Reserved
#
################################################################################
"""
This module provide The beginning of android mobile phone performance test

Authors:wangshuailong
Date:2022-12-12
"""
import sys
import os
import time
import requests


sys.path.append('../')


def get_token():
    try:
        with open('token.txt', encoding='utf-8') as file:
            token = file.read()
            return token
    except:
        print('请检查token是否正确,或token文件是否存在')
        sys.exit()


def get_rsa():
    try:
        url = 'http://perf.haimacloud.com/api/getpubkey/'
        with requests.post(url=url, data={'token': get_token()}) as response:
            key = response.json()['data']
        with open('pub.pem', 'w', newline='\n', encoding='utf-8') as f:
            f.write(key)
        print('用户信息验证通过')
    except:
        print('用户信息验证失败')


def push_romstat(args):
    push = 'adb -s ' + args + ' push romstat /data/local/tmp/'
    chmod = 'adb -s ' + args + ' shell chmod 777 /data/local/tmp/romstat'
    start = 'adb -s ' + args + ' shell /data/local/tmp/romstat -pem /data/local/tmp/pub.pem'
    push_rsa = 'adb -s ' + args + ' push pub.pem /data/local/tmp/'
    chmod_ras = 'adb -s ' + args + ' shell chmod 777 /data/local/tmp/pub.pem'
    os.system(push)
    os.system(chmod)
    os.system(push_rsa)
    os.system(chmod_ras)
    print('准备开始性能测试')
    time.sleep(2)
    os.system(start)


def main():
    r"""Perform performance test.

    :param ' ': If the parameter is empty, use the default android devices id.
    :param devices: Use the specified android devices id, only one can be specified.

    """
    try:
        push_romstat(sys.argv[1])
    except:
        rt = os.popen('adb devices').readlines()
        n = len(rt) - 2
        if n > 1:
            print('当前链接多台android设备，请指定需要上传的设备ID')
        else:
            for i in range(n):
                et = rt
                nPos = et[i + 1]
                dev = nPos[0: -7]
                print('当前链接设备号:' + dev)
                get_rsa()
                push_romstat(dev)


if __name__ == '__main__':
    main()
