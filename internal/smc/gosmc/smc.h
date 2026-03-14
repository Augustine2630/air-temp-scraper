/*
 * Apple System Management Control (SMC) Tool
 * Copyright (C) 2006 devnull
 * Adapted from https://github.com/dkorunic/iSMC
 *
 * Licensed under GPLv2+
 */

#include <IOKit/IOKitLib.h>

#if (MAC_OS_X_VERSION_MAX_ALLOWED < 120000)
	#define IOMainPort IOMasterPort
#endif

#ifndef __SMC_H__
#define __SMC_H__

#define KERNEL_INDEX_SMC      2

#define SMC_CMD_READ_BYTES    5
#define SMC_CMD_WRITE_BYTES   6
#define SMC_CMD_READ_INDEX    8
#define SMC_CMD_READ_KEYINFO  9
#define SMC_CMD_READ_PLIMIT   11
#define SMC_CMD_READ_VERS     12

#define SMC_TYPE_FPE2   "fpe2"
#define SMC_TYPE_FP2E   "fp2e"
#define SMC_TYPE_FP4C   "fp4c"
#define SMC_TYPE_CH8    "ch8*"
#define SMC_TYPE_SP78   "sp78"
#define SMC_TYPE_SP4B   "sp4b"
#define SMC_TYPE_FP5B   "fp5b"
#define SMC_TYPE_FP88   "fp88"
#define SMC_TYPE_UI8    "ui8"
#define SMC_TYPE_UI16   "ui16"
#define SMC_TYPE_UI32   "ui32"
#define SMC_TYPE_SI8    "si8"
#define SMC_TYPE_SI16   "si16"
#define SMC_TYPE_SI32   "si32"
#define SMC_TYPE_FLAG   "flag"
#define SMC_TYPE_FDS    "{fds"
#define SMC_TYPE_FLT    "flt"

typedef struct {
    UInt8  major;
    UInt8  minor;
    UInt8  build;
    UInt8  reserved[1];
    UInt16 release;
} SMCKeyData_vers_t;

typedef struct {
    UInt16 version;
    UInt16 length;
    UInt32 cpuPLimit;
    UInt32 gpuPLimit;
    UInt32 memPLimit;
} SMCKeyData_pLimitData_t;

typedef struct {
    UInt32 dataSize;
    UInt32 dataType;
    UInt8  dataAttributes;
} SMCKeyData_keyInfo_t;

typedef UInt8 SMCBytes_t[32];

typedef struct {
    UInt32                  key;
    SMCKeyData_vers_t       vers;
    SMCKeyData_pLimitData_t pLimitData;
    SMCKeyData_keyInfo_t    keyInfo;
    UInt8                   result;
    UInt8                   status;
    UInt8                   data8;
    UInt32                  data32;
    SMCBytes_t              bytes;
} SMCKeyData_t;

typedef char UInt32Char_t[5];

typedef struct {
    UInt32Char_t key;
    UInt32       dataSize;
    UInt32Char_t dataType;
    SMCBytes_t   bytes;
} SMCVal_t;

UInt32        _strtoul(const char *str, int size, int base);
void          _ultostr(char *str, UInt32 val);
kern_return_t SMCOpen(const char *serviceName, io_connect_t *conn);
kern_return_t SMCClose(io_connect_t conn);
kern_return_t SMCCall(io_connect_t conn, int index, SMCKeyData_t *inputStructure, SMCKeyData_t *outputStructure);
kern_return_t SMCReadKey(io_connect_t conn, const UInt32Char_t key, SMCVal_t *val);

#endif
