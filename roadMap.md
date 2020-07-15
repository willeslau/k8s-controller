# Overall road map:
1. (Done) Genereate controller code
2. (Done) Create/Delete `worker` and create/delete `deployment`
3. (Done) Update `worker` and update `deployment`
      (TODO) When worker spec is updated, the status does not reflect the change
4. () Watch for `deployment` changes and check against `worker`

Study `deployment-controller` in k8s source code, over there, they have adopt/release/orphan related processes, spend some time to study them.