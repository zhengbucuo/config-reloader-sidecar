## config-reloader-sidecar
为热加载sidecar容器
```yaml
kind: Deployment
apiVersion: apps/v1
...
spec:
  template:
    ...
    spec:
      volumes:
        - name: nginx-config
          configMap:
            name: nginx-conifg
            defaultMode: 420
      containers:
        - name: nginx
          image: nginx
          volumeMounts:
            - name: nginx-config
              readOnly: true
              mountPath: /etc/nginx/conf.d
         ...
        - name: nginx-reloader
          image: nginx-reloader
          env:
            - name: CONFIG_DIR
              value: /etc/nginx/conf.d/
            - name: PROCESS_NAME
              value: nginx
          volumeMounts:
            - name: nginx-config
              readOnly: true
              mountPath: /etc/nginx/conf.d
```
### 更新时间
更新速度默认为1分钟，
### 配置共享命名空间
为了让 sidecar 找到向哪个进程发送信号，需要将 Pod 配置为与共享进程命名空间shareProcessNamespace: true
