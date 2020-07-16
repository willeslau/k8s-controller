# Overall road map:
1. (Done) Genereate controller code
2. (Done) Create/Delete `worker` and create/delete `deployment`
3. (Done) Update `worker` and update `deployment`
4. (Done) Watch for `deployment` changes and check against `worker`
5. (In Progress) Setup image for RPC server
6. () Auto generate config map

Study `deployment-controller` in k8s source code, over there, they have adopt/release/orphan related processes, spend some time to study them.

Test: the efficiency of this controller. How many cycles does it take to complete update?