/*
 * Apple System Management Control (SMC) Tool
 * Copyright (C) 2006 devnull
 * Adapted from https://github.com/dkorunic/iSMC
 *
 * Licensed under GPLv2+
 */

#include <stdio.h>
#include <IOKit/IOKitLib.h>
#include <Kernel/string.h>
#include <os/lock.h>

#include "smc.h"

/* Cache keyInfo to reduce IOKit round-trips (energy-efficient). */
#define KEY_INFO_CACHE_SIZE 100
struct {
    UInt32               key;
    SMCKeyData_keyInfo_t keyInfo;
} g_keyInfoCache[KEY_INFO_CACHE_SIZE];

int          g_keyInfoCacheCount = 0;
os_unfair_lock g_keyInfoSpinLock = OS_UNFAIR_LOCK_INIT;

UInt32 _strtoul(const char *str, int size, int base)
{
    UInt32 total = 0;
    for (int i = 0; i < size; i++) {
        if (base == 16)
            total += str[i] << (size - 1 - i) * 8;
        else
            total += (unsigned char)(str[i] << (size - 1 - i) * 8);
    }
    return total;
}

void _ultostr(char *str, UInt32 val)
{
    str[4] = '\0';
    snprintf(str, 5, "%c%c%c%c",
             (unsigned int)val >> 24,
             (unsigned int)val >> 16,
             (unsigned int)val >> 8,
             (unsigned int)val);
}

kern_return_t SMCOpen(const char *serviceName, io_connect_t *conn)
{
    kern_return_t  result;
    mach_port_t    masterPort;
    io_iterator_t  iterator;
    io_object_t    device;

    IOMainPort(MACH_PORT_NULL, &masterPort);

    CFMutableDictionaryRef matchingDictionary = IOServiceMatching(serviceName);
    result = IOServiceGetMatchingServices(masterPort, matchingDictionary, &iterator);
    if (result != kIOReturnSuccess)
        return 1;

    device = IOIteratorNext(iterator);
    IOObjectRelease((io_object_t)iterator);
    if (device == 0)
        return 1;

    result = IOServiceOpen(device, mach_task_self(), 0, conn);
    IOObjectRelease(device);
    if (result != kIOReturnSuccess)
        return 1;

    return kIOReturnSuccess;
}

kern_return_t SMCClose(io_connect_t conn)
{
    return IOServiceClose(conn);
}

kern_return_t SMCCall(io_connect_t conn, int index,
                      SMCKeyData_t *inputStructure,
                      SMCKeyData_t *outputStructure)
{
    size_t structureInputSize  = sizeof(SMCKeyData_t);
    size_t structureOutputSize = sizeof(SMCKeyData_t);

    return IOConnectCallStructMethod(conn, index,
                                     inputStructure,  structureInputSize,
                                     outputStructure, &structureOutputSize);
}

static kern_return_t SMCGetKeyInfo(io_connect_t conn, UInt32 key,
                                   SMCKeyData_keyInfo_t *keyInfo)
{
    SMCKeyData_t  inputStructure;
    SMCKeyData_t  outputStructure;
    kern_return_t result = kIOReturnSuccess;
    int i = 0;

    os_unfair_lock_lock(&g_keyInfoSpinLock);

    for (; i < g_keyInfoCacheCount; ++i) {
        if (key == g_keyInfoCache[i].key) {
            *keyInfo = g_keyInfoCache[i].keyInfo;
            break;
        }
    }

    if (i == g_keyInfoCacheCount) {
        memset(&inputStructure,  0, sizeof(inputStructure));
        memset(&outputStructure, 0, sizeof(outputStructure));

        inputStructure.key   = key;
        inputStructure.data8 = SMC_CMD_READ_KEYINFO;

        result = SMCCall(conn, KERNEL_INDEX_SMC, &inputStructure, &outputStructure);
        if (result == kIOReturnSuccess) {
            *keyInfo = outputStructure.keyInfo;
            if (g_keyInfoCacheCount < KEY_INFO_CACHE_SIZE) {
                g_keyInfoCache[g_keyInfoCacheCount].key     = key;
                g_keyInfoCache[g_keyInfoCacheCount].keyInfo = outputStructure.keyInfo;
                ++g_keyInfoCacheCount;
            }
        }
    }

    os_unfair_lock_unlock(&g_keyInfoSpinLock);
    return result;
}

kern_return_t SMCReadKey(io_connect_t conn, const UInt32Char_t key, SMCVal_t *val)
{
    kern_return_t result;
    SMCKeyData_t  inputStructure;
    SMCKeyData_t  outputStructure;

    memset(&inputStructure,  0, sizeof(SMCKeyData_t));
    memset(&outputStructure, 0, sizeof(SMCKeyData_t));
    memset(val,              0, sizeof(SMCVal_t));

    inputStructure.key = _strtoul(key, 4, 16);
    memcpy(val->key, key, sizeof(val->key));

    result = SMCGetKeyInfo(conn, inputStructure.key, &outputStructure.keyInfo);
    if (result != kIOReturnSuccess)
        return result;

    val->dataSize = outputStructure.keyInfo.dataSize;
    _ultostr(val->dataType, outputStructure.keyInfo.dataType);
    inputStructure.keyInfo.dataSize = val->dataSize;
    inputStructure.data8            = SMC_CMD_READ_BYTES;

    result = SMCCall(conn, KERNEL_INDEX_SMC, &inputStructure, &outputStructure);
    if (result != kIOReturnSuccess)
        return result;

    memcpy(val->bytes, outputStructure.bytes, sizeof(outputStructure.bytes));
    return kIOReturnSuccess;
}
