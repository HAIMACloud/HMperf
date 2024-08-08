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
import json

t = time.time()

def get_token():

    try:
        with open("./token.txt", encoding='utf-8') as file:
            token = file.read()
            return token
    except:
        print('请检查token是否正确,或token文件是否存在')
        sys.exit()

def up_report():
    url = 'http://perf.haimacloud.com/api/uploaddatasource/'
    token = get_token()
    testername = get_token()
    path = './out.hmp'
    devicesId = 'haimayun'
    devicesNname = 'windows'
    packageName = 'x86.test'
    systemVersion = 'windows'
    system_os = 'windows'  
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

if __name__ == '__main__':
    up_report()
