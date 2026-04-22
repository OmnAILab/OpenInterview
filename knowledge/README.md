# 面试知识库模板

这个目录是一套适合中文面试场景的 RAG 知识库模板。

服务端会把 `.md` 文件切块，并在每次回答前检索相关片段注入到提示词里。所以这里的内容应该尽量真实、简洁、可检索。

## 如何使用

在 `.env` 中设置：

```bash
INTERVIEW_KNOWLEDGE_LOCAL_PATH=./knowledge
```

然后把这些模板替换成你的真实经历。建议优先补充以下内容：

- 你做过什么
- 业务背景是什么
- 你具体负责什么
- 做决策时有哪些取舍
- 最终结果和指标是什么
- 遇到过哪些问题，复盘后学到了什么

不要写：

- 密钥、令牌、客户隐私、商业机密
- “优化了很多”“效果很好” 这类没有证据的描述
- 把团队成果全部说成你个人成果
- 自己没做过却假装做过的经历

## 为了让 RAG 更容易命中

尽量直接使用中文面试官会说的话。

例如：

- 请你做一下自我介绍
- 介绍一下你的项目
- 你为什么想来我们公司
- 讲一个你最有挑战的项目
- 你做过高并发吗
- 你怎么做系统设计
- 你遇到过线上故障吗
- 你是怎么排查问题的
- 你最大的优点和缺点是什么

建议一个主题一个章节，一个项目一个文件。如果你的项目或故事比较多，可以复制对应模板并重命名。

## 建议填写顺序

1. `01_candidate_profile.md`
2. `02_self_intro.md`
3. `03_role_fit.md`
4. `projects/01_primary_project.md`
5. `projects/02_secondary_project.md`
6. `stories/01_behavioral_stories.md`
7. `technical/01_system_design_topics.md`
8. `technical/02_incident_debugging.md`
9. `company/01_target_company_research.md`
10. `company/02_reverse_questions.md`
