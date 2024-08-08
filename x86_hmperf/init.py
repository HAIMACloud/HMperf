################################################################################
#
# Copyright (c) 2022 haimacloud.com, Inc. All Rights Reserved
#
################################################################################
"""
This module provide Upload of test results

Authors:wangshuailong
Date:2023-11-17
"""

import sys
import os
import requests

sys.path.append('../')
__current_path = os.path.dirname(os.path.realpath(__file__))
print(__current_path)

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

if __name__ == '__main__':
    get_rsa()