#!/bin/bash
# ============================================================================
# install-docker.sh - Docker-specific functions (sourced, not executed)
# ----------------------------------------------------------------------------
# This file contains the entire Docker installation logic extracted from
# the original install.sh. Source it from scripts that need Docker ops:
#
#   source "${SCRIPT_DIR}/install-lib.sh"
#   source "${SCRIPT_DIR}/install-docker.sh"
# ============================================================================

# ----------------------------------------------------------------------------
# Mirror selection — pick the fastest Docker CE repo
# ----------------------------------------------------------------------------
Docker_Download_Url_Check() {
    echo -e "Selecting download docker address source, please wait..."
    local ser_names=(download.docker.com mirrors.ustc.edu.cn/docker-ce \
                     mirrors.tuna.tsinghua.edu.cn/docker-ce \
                     mirrors.aliyun.com/docker-ce mirror.azure.cn/docker-ce)
    local tmp_file1=/dev/shm/net_test1.pl
    local tmp_file2=/dev/shm/net_test2.pl
    [ -f "${tmp_file1}" ] && rm -f "${tmp_file1}"
    [ -f "${tmp_file2}" ] && rm -f "${tmp_file2}"
    touch "${tmp_file1}"
    touch "${tmp_file2}"

    local ser_name NODE_CHECK NODE_STATUS TIME_TOTAL
    for ser_name in "${ser_names[@]}"; do
        NODE_CHECK=$(curl -k -s 2>/dev/null --connect-timeout 5 \
            -w "%{http_code} %{time_total}" "https://${ser_name}" \
            -o "c${ser_name}.txt" 2>/dev/null | xargs)
        rm -rf "c${ser_name}.txt"
        NODE_STATUS=$(echo "${NODE_CHECK}" | awk '{print $1}')
        TIME_TOTAL=$(echo "${NODE_CHECK}" | awk '{print $2 * 1000 - 500}' | cut -d '.' -f 1)
        if [ "${NODE_STATUS}" == "200" ] || [ "${NODE_STATUS}" == "301" ] || [ "${NODE_STATUS}" == "403" ]; then
            if [ "${TIME_TOTAL:-99999}" -lt 100 ] 2>/dev/null; then
                echo "${ser_name}" >> "${tmp_file1}"
            elif [ "${TIME_TOTAL:-0}" -gt 1500 ] && [ "${TIME_TOTAL:-0}" -lt 2500 ] 2>/dev/null; then
                echo "${ser_name}" >> "${tmp_file2}"
            fi
        fi
    done
    Docker_NODE_URL=$(cat "${tmp_file1}" 2>/dev/null | head -n 1 | awk '{print $1}')
    if [ -z "${Docker_NODE_URL}" ]; then
        Docker_NODE_URL=$(cat "${tmp_file2}" 2>/dev/null | head -n 1 | awk '{print $1}')
    fi
    rm -f "${tmp_file1}" "${tmp_file2}"
    mirror="${Docker_NODE_URL}"
    echo "${mirror}"

    case "${mirror}" in
        mirrors.aliyun.com/docker-ce)
            Default_Download_Url="https://mirrors.aliyun.com/docker-ce"
            mirror="Aliyun"
            ;;
        mirror.azure.cn/docker-ce)
            Default_Download_Url="https://mirror.azure.cn/docker-ce"
            mirror="AzureChinaCloud"
            ;;
        mirrors.ustc.edu.cn/docker-ce)
            Default_Download_Url="https://mirrors.ustc.edu.cn/docker-ce"
            mirror="USTC"
            ;;
        mirrors.tuna.tsinghua.edu.cn/docker-ce)
            Default_Download_Url="https://mirrors.tuna.tsinghua.edu.cn/docker-ce"
            mirror="tsinghua"
            ;;
        *)
            Default_Download_Url="https://download.docker.com"
            mirror=""
            ;;
    esac

    DEFAULT_REPO_FILE="docker-ce.repo"
    [ -z "${REPO_FILE}" ] && REPO_FILE="${DEFAULT_REPO_FILE}"
}

# ----------------------------------------------------------------------------
# Main Docker CE installer (apt + yum paths)
# ----------------------------------------------------------------------------
Docker_Install() {
    local lsb_dist lsb_name dist_version
    lsb_dist=$(Get_Distribution | awk -F " " '{print $1}')
    lsb_dist="$(echo "${lsb_dist}" | tr '[:upper:]' '[:lower:]')"

    case "${lsb_dist}" in
        ubuntu)
            if Command_Exists lsb_release; then
                dist_version="$(lsb_release --codename | cut -f2)"
            fi
            if [ -z "${dist_version}" ] && [ -r /etc/lsb-release ]; then
                dist_version="$(. /etc/lsb-release && echo "$DISTRIB_CODENAME")"
            fi
            ;;
        debian | raspbian)
            dist_version="$(sed 's/\/.*//' /etc/debian_version | sed 's/\..*//')"
            case "${dist_version}" in
                12) dist_version="bookworm" ;;
                11) dist_version="bullseye" ;;
                10) dist_version="buster" ;;
                9)  dist_version="stretch" ;;
                8)  dist_version="jessie" ;;
            esac
            ;;
        centos | rhel | sles | ol | tencentos | alinux | anolis | rocky | euleros | almalinux | opencloudos | openeuler | hce | kylin | uos | kylinsecos | amzn)
            if [ -z "${dist_version}" ] && [ -r /etc/os-release ]; then
                dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
            fi
            ;;
        *)
            if Command_Exists lsb_release; then
                dist_version="$(lsb_release --release | cut -f2)"
            fi
            if [ -z "${dist_version}" ] && [ -r /etc/os-release ]; then
                dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
            fi
            ;;
    esac

    Check_Forked

    local conflicting=""
    case "${lsb_dist}" in
        ubuntu | debian | raspbian)
            local docker_gpg="/usr/share/keyrings/docker-archive-keyring.gpg"
            local apt_repo_file="/etc/apt/sources.list.d/docker.list"
            local pre_reqs="apt-transport-https ca-certificates curl wget"
            command -v gpg >/dev/null || pre_reqs="${pre_reqs} gnupg"

            local apt_repo="deb [arch=$(dpkg --print-architecture) signed-by=${docker_gpg}] ${Default_Download_Url}/linux/${lsb_dist} ${dist_version} stable"
            (
                DEBIAN_FRONTEND=noninteractive apt-get install -y ${pre_reqs}
                if [ -f "${docker_gpg}" ] && [ -f "${apt_repo_file}" ]; then
                    rm -rf "${docker_gpg}"
                    rm -rf "${apt_repo_file}"
                fi
                wget -qO - -t 5 -T 10 "${Default_Download_Url}/linux/${lsb_dist}/gpg" \
                    | gpg --dearmor --yes -o "${docker_gpg}"
                chmod a+r "${docker_gpg}"
                echo "${apt_repo}" > "${apt_repo_file}"
                apt-get update -y
            )
            (
                DEBIAN_FRONTEND=noninteractive apt-get install -y \
                    docker-ce docker-ce-cli containerd.io \
                    docker-buildx-plugin docker-compose-plugin
            )
            ;;

        centos | fedora | rhel | ol | tencentos | alinux | anolis | rocky | almalinux | opencloudos | openeuler | hce | uos | amzn)
            local pkg_version="" pkg_manager yum_repo
            yum_repo="${Default_Download_Url}/linux/${lsb_dist}/${REPO_FILE}"
            if [ "${lsb_dist}" == "ol" ] || [ "${lsb_dist}" == "tencentos" ] || \
               [ "${lsb_dist}" == "alinux" ] || [ "${lsb_dist}" == "anolis" ] || \
               [ "${lsb_dist}" == "opencloudos" ] || [ "${lsb_dist}" == "openeuler" ] || \
               [ "${lsb_dist}" == "hce" ] || [ "${lsb_dist}" == "uos" ] || \
               [ "${lsb_dist}" == "amzn" ]; then
                yum_repo="${Default_Download_Url}/linux/centos/${REPO_FILE}"
            fi
            if [ "${lsb_dist}" == "rocky" ] || [ "${lsb_dist}" == "almalinux" ]; then
                yum_repo="${Default_Download_Url}/linux/centos/${REPO_FILE}"
            fi

            local pre_reqs pkg_suffix config_manager enable_channel_flag disable_channel_flag
            if [ "${lsb_dist}" = "fedora" ]; then
                pkg_manager="dnf"
                config_manager="dnf config-manager"
                enable_channel_flag="--set-enabled"
                disable_channel_flag="--set-disabled"
                pre_reqs="dnf-plugins-core"
                pkg_suffix="fc${dist_version}"
            else
                pkg_manager="yum"
                config_manager="yum-config-manager"
                enable_channel_flag="--enable"
                disable_channel_flag="--disable"
                pre_reqs="yum-utils"
                pkg_suffix="el"
            fi

            wget -O /etc/yum.repos.d/docker-ce.repo "${yum_repo}" -t 5 -T 10
            if [ $? -ne "0" ] || ! grep -q "docker" /etc/yum.repos.d/docker-ce.repo; then
                create_docker_yum_repo
            fi
            sed -i "s|https://download.docker.com|${Default_Download_Url}|g" /etc/yum.repos.d/docker-ce.repo

            lsb_name=$(Get_Distribution | awk -F " " '{print $3}')
            [ "${lsb_name}" = "Stream" ] && conflicting="--allowerasing"
            [ "${lsb_name}" = "rocky" ] || [ "${lsb_dist}" = "euleros" ] || [ "${lsb_dist}" = "almalinux" ] && conflicting="--allowerasing"

            if [ "${lsb_dist}" = "openeuler" ]; then
                local openeuler_v
                openeuler_v=$(cat /etc/os-release | grep "VERSION=")
                if [[ "${openeuler_v}" =~ "23.0" ]]; then
                    sed -i "s|\$releasever|8|g" /etc/yum.repos.d/docker-ce.repo
                    ${pkg_manager} makecache
                elif [[ "${openeuler_v}" =~ "22.0" ]]; then
                    sed -i "s|\$releasever|7|g" /etc/yum.repos.d/docker-ce.repo
                    ${pkg_manager} makecache
                else
                    [ -f "/etc/os-release" ] && cat /etc/os-release
                    log_error "not support this system yum auto install docker"
                fi
            fi

            if [ "${lsb_dist}" = "amzn" ]; then
                sed -i "s|\$releasever|8|g" /etc/yum.repos.d/docker-ce.repo
                ${pkg_manager} makecache
            fi

            if [ "${lsb_dist}" = "hce" ]; then
                local hce_v
                hce_v=$(. /etc/os-release && echo "$VERSION_ID")
                if [[ "${hce_v}" =~ "2.0" ]]; then
                    sed -i "s|\$releasever|8|g" /etc/yum.repos.d/docker-ce.repo
                    ${pkg_manager} makecache
                fi
            fi

            ${pkg_manager} makecache
            ${pkg_manager} install -y \
                docker-ce${pkg_version} docker-ce-cli containerd.io \
                docker-buildx-plugin docker-compose-plugin ${conflicting}
            ;;

        *)
            if [ -z "${lsb_dist}" ]; then
                if Is_Darwin; then
                    log_error "Unsupported operating system 'macOS'"
                    exit 1
                fi
            fi
            Red_Error "ERROR: not support '${lsb_dist}' this system auto install docker"
            ;;
    esac

    # Fallback: binary tarball install
    if ! Command_Exists docker && [ ! -f /usr/bin/docker ]; then
        log_warn "Trying binary package install..."
        Binary_package_install_docker
    fi

    if ! Command_Exists docker; then
        Red_Error "docker command not exists, install docker failed."
    fi
}

# ----------------------------------------------------------------------------
# Fallback: install Docker via static binary tarball
# ----------------------------------------------------------------------------
Binary_package_install_docker() {
    local DOCKER_VERSION="27.3.1"
    local SYS_TYPE
    SYS_TYPE=$(uname -m)

    log_info "Create docker group"
    groupadd docker 2>/dev/null || true

    log_info "Downloading docker ${DOCKER_VERSION}.tgz"
    wget --no-check-certificate -O "docker-${DOCKER_VERSION}.tgz" \
        "${DOWNLOAD_URL}/src/docker-${DOCKER_VERSION}-${SYS_TYPE}.tgz" -t 5 -T 10

    local tmp_size
    tmp_size=$(du -b "docker-${DOCKER_VERSION}.tgz" | awk '{print $1}')
    if [ "${tmp_size:-0}" -lt 60831072 ]; then
        ls -l "docker-${DOCKER_VERSION}.tgz"
        Red_Error "Failed to download docker tgz. Please retry."
    fi

    log_info "Extracting docker binaries"
    tar -xvf "docker-${DOCKER_VERSION}.tgz"
    \cp docker/* /usr/bin/ 2>/dev/null || cp docker/* /usr/bin/
    rm -f "docker-${DOCKER_VERSION}.tgz"

    wget --no-check-certificate -O /usr/bin/docker-compose \
        "${DOWNLOAD_URL}/src/docker-compose-linux-${SYS_TYPE}" -t 5 -T 20
    chmod +x /usr/bin/docker-compose

    wget --no-check-certificate -O /usr/lib/systemd/system/docker.service \
        "${DOWNLOAD_URL}/init/systemd/docker.service" -t 5 -T 20
    wget --no-check-certificate -O /usr/lib/systemd/system/docker.socket \
        "${DOWNLOAD_URL}/init/systemd/docker.socket" -t 5 -T 20
    wget --no-check-certificate -O /usr/lib/systemd/system/containerd.service \
        "${DOWNLOAD_URL}/init/systemd/containerd.service" -t 5 -T 20

    systemctl enable docker.service
    systemctl enable docker.socket
    systemctl enable containerd.service
    systemctl start containerd.service
    systemctl start docker.socket
    systemctl start docker.service
}

# ----------------------------------------------------------------------------
# Install Docker Compose (CLI plugin or standalone)
# ----------------------------------------------------------------------------
Docker_Compose_Install() {
    if [ "${is_docker_compose}" == "0" ]; then
        Docker_Install
    fi

    local Compose_Download_Url="${DOWNLOAD_URL}/install/src/docker-compose-$(uname -s)-$(uname -m)"
    local Compose_Path="/usr/local/bin/docker-compose"
    local New_Compose_Path="/usr/libexec/docker/cli-plugins/docker-compose"
    local Compose_lin="/usr/bin/docker-compose"

    if [ ! -f "${New_Compose_Path}" ]; then
        wget --no-check-certificate -t 5 -T 20 -O "${New_Compose_Path}" "${Compose_Download_Url}"
        chmod +x "${New_Compose_Path}"
        rm -rf "${Compose_lin}"
    fi

    if [ ! -f "${Compose_lin}" ]; then
        ln -s "${New_Compose_Path}" "${Compose_lin}"
    fi
}

# ----------------------------------------------------------------------------
# Start docker.service
# ----------------------------------------------------------------------------
Docker_Start() {
    if Command_Exists docker || [ -f /lib/systemd/system/docker.service ]; then
        if Command_Exists systemctl; then
            grep docker /etc/group >/dev/null 2>&1 || groupadd docker
            systemctl enable docker
            systemctl reset-failed docker.service
            systemctl start docker.service
            if [ "$?" != "0" ]; then
                systemctl reset-failed docker.service
                systemctl restart docker.service
                if [ "$?" != "0" ]; then
                    local docker_status
                    docker_status=$(systemctl status docker.service)
                    echo "${docker_status}"
                    if [ -f /etc/docker/daemon.json ]; then
                        echo "/etc/docker/daemon.json content:"
                        cat /etc/docker/daemon.json
                        log_warn "Check /etc/docker/daemon.json syntax."
                    fi
                    log_error "docker start failed, retrying in 100s..."
                    Bored_waiting 100 "waiting to start docker"
                    systemctl reset-failed docker.service
                    systemctl restart docker.service
                    if [ "$?" != "0" ]; then
                        docker_status=$(systemctl status docker.service)
                        echo "${docker_status}"
                        Red_Error "docker start failed" \
                                  "Try: systemctl restart docker.service" \
                                  "Then re-run the installer."
                    fi
                fi
            fi
        else
            chkconfig --add docker
            chkconfig --level 2345 docker on
            service docker start
            if [ "$?" != "0" ]; then
                service docker.service status
                cat /etc/docker/daemon.json 2>/dev/null || true
                Red_Error "docker start failed"
            fi
        fi
    else
        Red_Error "Docker is not installed"
    fi
}
