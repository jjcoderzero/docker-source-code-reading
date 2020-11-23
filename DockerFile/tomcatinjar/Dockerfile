#From ic-harbor.baozun.com/ic/alpine-jdk1.8:latest
FROM ic-harbor.baozun.com/skywalking/lune-agent-openjdk8:2.3.212

USER vmuser
RUN mkdir /opt/project/ -p
WORKDIR /opt/project
ADD ./apple-adapter-impl/docker/startup-sw.sh /opt/project
#打包后项目名（有pom.xml文件中finalName）
ADD ./apple-adapter-impl/target/apple-adapter-impl.jar /opt/project
ENTRYPOINT ["sh", "startup-sw.sh", "apple-adapter-impl"]
