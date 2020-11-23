#!/bin/sh

sleep 2
if !(which java 2>/dev/null); then
    echo '请安装java环境'
    exit
fi
  
PROJNAME=${PROJNAME} 
APPNAME=${APPNAME}
 
PROJ_APP=${PROJNAME}\_${APPNAME}
if [ ${#PROJ_APP} > 24 ];then
    PROJ_APP=${PROJ_APP:0:24}
fi
 
PROJECT_NAME="$1"
 
agentID=$POD_IP
envID=$SPRING_PROFILES_ACTIVE
 
MEM_OPTS="-Xms2g -Xmx2g -Xmn768m"
 
PINPOINT_OPS=""
PINPOINT_ENABLED=${PINPOINT_ENABLED:-"false"}
PINPOINT_IP=${PINPOINT_IP:-"127.0.0.1"}
PROFILER_SAMPLING_RATE=${PROFILER_SAMPLING_RATE:-10}
if [ "$PINPOINT_ENABLED" == "true" ]; then
   sed -i "/profiler.collector.ip=/ s/=.*/=${PINPOINT_IP}/" /data/pp-agent/pinpoint.config
   sed -i "/profiler.sampling.rate=/ s/=.*/=${PROFILER_SAMPLING_RATE}/" /data/pp-agent/pinpoint.config
   PINPOINT_OPS="-javaagent:/data/pp-agent/pinpoint-bootstrap.jar -Dpinpoint.container -Dpinpoint.agentId=${agentID} -Dpinpoint.applicationName=$PROJ_APP"
fi
 
# java $MEM_OPTS $GC_OPTS $JMX_OPTS $START_OPTS $PINPOINT_OPS -jar -Dspring.profiles.active=${envID} -server ${PROJECT_NAME}.jar
 
nohup java $PINPOINT_OPS $MEM_OPTS -jar -Dspring.profiles.active=${envID} -server ${PROJECT_NAME}.jar
