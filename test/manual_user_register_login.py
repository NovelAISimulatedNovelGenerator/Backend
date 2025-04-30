# -*- coding: utf-8 -*-
"""
manual_user_register_login.py
手动测试用户注册和登录接口，便于排查自动化测试失败原因。
"""
import requests
import json

BASE_URL = "http://127.0.0.1:8888"
REGISTER_URL = f"{BASE_URL}/api/user/register"
LOGIN_URL = f"{BASE_URL}/api/user/login"

USERNAME = "testuser_new"
PASSWORD = "testpass"

if __name__ == "__main__":
    print("------ 注册用户 ------")
    reg_resp = requests.post(REGISTER_URL, json={"username": USERNAME, "password": PASSWORD})
    print(f"注册 status: {reg_resp.status_code}")
    try:
        print("注册响应:", json.dumps(reg_resp.json(), ensure_ascii=False, indent=2))
    except Exception:
        print("注册响应非 json:", reg_resp.text)

    print("\n------ 登录用户 ------")
    login_resp = requests.post(LOGIN_URL, json={"username": USERNAME, "password": PASSWORD})
    print(f"登录 status: {login_resp.status_code}")
    try:
        print("登录响应:", json.dumps(login_resp.json(), ensure_ascii=False, indent=2))
    except Exception:
        print("登录响应非 json:", login_resp.text)
