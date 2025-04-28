# -*- coding: utf-8 -*-
"""
test_user_delete.py
删除用户接口（/api/user/delete）API单元测试

- 覆盖正常删除、未登录、token异常、重复删除等情况
- 需先启动后端服务
- 依赖 pytest fixture auto_register_and_login
"""
import requests
import pytest

BASE_URL = "http://127.0.0.1:8888"
DELETE_URL = f"{BASE_URL}/api/user/delete"

class TestUserDelete:
    """
    /api/user/delete 删除用户接口测试
    """
    def test_delete_user_success(self, auto_register_and_login):
        """
        正常删除用户（删除后token失效，需注意幂等）
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.delete(DELETE_URL, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert "删除成功" in data["message"]

    def test_delete_user_unauthorized(self):
        """
        未携带token，鉴权失败
        """
        resp = requests.delete(DELETE_URL)
        assert resp.status_code in (401, 403)

    def test_delete_user_invalid_token(self):
        """
        携带无效token，鉴权失败
        """
        headers = {"Authorization": "Bearer invalidtoken123"}
        resp = requests.delete(DELETE_URL, headers=headers)
        assert resp.status_code in (401, 403)

    def test_delete_user_twice(self, auto_register_and_login):
        """
        重复删除同一用户，应提示用户不存在或已删除
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        # 第一次删除
        resp1 = requests.delete(DELETE_URL, headers=headers)
        # 第二次删除
        resp2 = requests.delete(DELETE_URL, headers=headers)
        assert resp2.status_code == 200
        data = resp2.json()
        assert data["code"] != 200
        assert "不存在" in data["message"] or "已删除" in data["message"] or "失败" in data["message"]

# 运行方法：pytest test_user_delete.py
