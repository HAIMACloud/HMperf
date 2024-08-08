################################################################################
#
# Copyright (c) 2022 haimacloud.com, Inc. All Rights Reserved
#
################################################################################
"""
This module provide Upload of test results

Authors:wangshuailong
Date:2022-12-12
"""
import requests
import os
import time
import sys
import re
import platform
import json

t = time.time()


def res_expression(lb, rb, data):
    data = data
    rule = lb + r"(.*?)" + rb
    association = re.findall(rule, data)
    return association


def get_token():
    try:
        with open('token.txt', encoding='utf-8') as file:
            token = file.read()
            return token
    except:
        print('请检查token是否正确,或token文件是否存在')
        sys.exit()


def get_test_info(args):
    try:
        filename = 'out.hmp'
        os.system('adb -s ' + args + ' pull data/local/tmp/out.hmp ./')
        dev = args
        devices_name = os.popen('adb -s ' + args + ' shell getprop ro.product.model').readlines()
        android_SDK = int(os.popen('adb -s ' + args + ' shell getprop ro.build.version.sdk').readlines()[0])
        if android_SDK > 32:
            if platform.system().lower() == 'windows':
                package_name = res_expression(lb='u0 ', rb='/',
                                            data=str(os.popen(
                                                'adb -s ' + args + ' shell dumpsys activity | findstr  topResumedActivity').readlines()))
            else:
                package_name = res_expression(lb='u0 ', rb='/',
                                            data=str(os.popen(
                                                'adb -s ' + args + ' shell dumpsys activity | grep  topResumedActivity').readlines()))
        else:
            if platform.system().lower() == 'windows':
                package_name = res_expression(lb='u0 ', rb='/',
                                              data=str(
                                                  os.popen(
                                                      'adb -s ' + args + ' shell dumpsys window | findstr mCurrentFocus').readlines()))

            else:
                package_name = res_expression(lb='u0 ', rb='/',
                                              data=str(os.popen(
                                                  'adb -s ' + args + ' shell dumpsys activity | grep mResumedActivity').readlines()))
        android_vesion = os.popen('adb -s ' + args + ' shell getprop ro.build.version.release').readlines()
        return os.getcwd() + '/' + filename, dev.strip(), devices_name[0], package_name[0], android_vesion
    except:
        print('未连接手机或测试文件不存在,请检查')
        os.unlink(os.getcwd() + '/' + filename)
        sys.exit()


def up_report(args):
    url = 'http://perf.haimacloud.com/api/uploaddatasource/'
    test_info = get_test_info(args)
    token = get_token()
    testername = get_token()
    path = test_info[0]
    devicesId = test_info[1]
    devicesNname = test_info[2].strip()
    packageName = test_info[3]
    systemVersion = 'android' + ' ' + test_info[4][0].strip()
    system_os = 'android'
    testername = testername
    remarks = input('输入测试场景，点击回车确认:')
    data = {
        "testerName": testername,
        'testType': 2,
        'testTimes': int(t),
        'devicesId': devicesId,
        'devicesNname': devicesNname,
        'packageName': packageName,
        'systemVersion': systemVersion,
        'system_os': system_os,
        'remarks': remarks,
        'token': token
    }
    fp = open(path, 'rb')
    print(data)
    with requests.post(url, files={'file': fp}, data=data) as response:
        msg = json.loads(response.text)
        print(msg)
        if msg['errmsg'] == 'success':
            print('上传成功')
            fp.close()
            os.unlink(path)
        else:
            print('上传失败')
            print(response.text)


def main():
    r"""Upload test results.

    :param ' ': If the parameter is empty, use the default android devices id.
    :param devices: Use the specified android devices id, only one can be specified.

    """
    get_token()
    try:
        up_report(sys.argv[1])
    except:
        rt = os.popen('adb devices').readlines()
        n = len(rt) - 2
        if n == 0:
            print('当前未连接android设备，请检查设备是否连接或是否打开usb调试模式')
            sys.exit()
        if n > 1:
            print('当前链接多台android设备，请指定需要上传的设备ID')
        else:
            for i in range(n):
                et = rt
                nPos = et[i + 1]
                dev = nPos[0: -7]
                print('当前链接设备号:' + dev)
                up_report(dev)


if __name__ == '__main__':
    main()
