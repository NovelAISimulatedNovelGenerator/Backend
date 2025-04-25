# -*- coding: utf-8 -*-
"""
测试 /api/save/list 接口
"""
import pytest
import requests

LIST_SAVE_URL = "http://localhost:8888/api/save/list"
CREATE_SAVE_URL = "http://localhost:8888/api/save/create"

class TestSaveList:
    def create_save(self, token: str, name: str) -> str:
        payload = {
            "save_name": name,
            "save_data": "内容",
            "save_type": "novel",
            "save_description": "章节"
        }
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.post(CREATE_SAVE_URL, json=payload, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert "save_id" in data
        return data["save_id"]

    def test_list_save_success(self, auto_register_and_login):
        """
        正常分页列出存档
        """
        token = auto_register_and_login
        # 创建多条数据
        for i in range(5):
            self.create_save(token, f"list测试存档{i}")
        payload = {"page": 1, "page_size": 10}
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.get(LIST_SAVE_URL, params=payload, headers=headers)
        print(f"[test_list_save_success] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert data["message"] == "获取成功"
        assert isinstance(data["saves"], list)
        assert data["total"] >= 5

    def test_list_save_missing_params(self, auto_register_and_login):
        """
        缺少分页参数，应返回400
        """
        token = auto_register_and_login
        payload = {}
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.get(LIST_SAVE_URL, params=payload, headers=headers)
        print(f"[test_list_save_missing_params] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code == 400
        data = resp.json()
        assert data["code"] == 400
        assert "分页参数非法" in data["message"] or "请求参数不合法" in data["message"]

    def test_list_save_unauthorized(self):
        """
        未授权请求，应返回401
        """
        payload = {"page": 1, "page_size": 10}
        resp = requests.get(LIST_SAVE_URL, params=payload)
        print(f"[test_list_save_unauthorized] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code in (401, 403)
