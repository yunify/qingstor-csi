| 描述 |  命令 | 返回值 | 返回消息|
|:---:|:---:|:----:|:---:|
|查询存在的Volume info| list_volume -pool csi -volume foo --detail |0|Volume Count:  2... [TABLE]|
|查询不存在的Volume info|list_volume -pool csi -volume fake --detail |0|Volume Count:0|
|查询Pool不存在的Volume info|list_volume -pool fake -volume foo --detail |1|FATA[0000] Failed ...|
|查询存在的Pool info|stats_pool --pool csi|0|[TABLE]|
|查询不存在的Pool info|stats_pool --pool foo|1|[NULL]|
|查询存在的Pool名|list_pool -pool csi -detail|0|Pool Count:  1... [TABLE]|
|查询不存在的Pool名|list_pool -pool fake -detail|0|Pool Count:0|
|查询存在的Snapshot info|list_snapshot --volume foo --pool csi|0|Snapshot Count:  1... [TABLE]|
|查询不存在snapshot的Snapshot info|list_snapshot --volume foo --pool csi|0|snapshot count:0|
|查询不存在volume的Snapshot info|list_snapshot --volume fake --pool csi|1|FATA[0000] Failed to list snapshot on volume:fake, reason:HTTP status:400  rc:-102 reason:Volume not exists|
|查询不存在pool的Snapshot info|list_snapshot --volume foo --pool fake|1|FATA[0000] Failed to list snapshot on volume:foo, reason:HTTP status:400  rc:-102 reason:Volume not exists|

