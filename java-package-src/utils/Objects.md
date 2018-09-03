---
title: Objects
categories:
  - java-package-src
  - utils
---



## src

```java
package java.util;

import java.util.function.Supplier;

/**
 * Objects是一个静态工具类，提供了带有null检测和不带有null检测的计算hash值，获得对象
 * 字符串，比较两个对象的功能
 */
public final class Objects {
    private Objects() { // 关闭构造方法
        throw new AssertionError("No java.util.Objects instances for you!");
    }

    /**
     * 返回两个对象是否相等
     * 如果两个对象都为null，返回true
     * 如果第一个参数不为null，调用a.equals()来返回值
     */
    public static boolean equals(Object a, Object b) {
        return (a == b) || (a != null && a.equals(b));
    }

   /**
    * 返回 deeplyEqual 的比较结果
    */
    public static boolean deepEquals(Object a, Object b) {
        if (a == b) // 包含了都为null的情况
            return true;
        else if (a == null || b == null)
            return false;
        else
            return Arrays.deepEquals0(a, b); // Q2A Arrays.deepEquals0(a,b)
    }

    /**
     * 返回对象hash值，如果为null返回0
     */
    public static int hashCode(Object o) {
        return o != null ? o.hashCode() : 0;
    }

   /**
    * 返回多个对象的hash值
    * 最佳实践是计算一个类，包含多个域的类，的hash值
    */
    public static int hash(Object... values) {
        return Arrays.hashCode(values);
    }

    /**
     * 返回对象的字符串值，对于非null返回字符串，对于null返回null
     */
    public static String toString(Object o) {
        return String.valueOf(o);
    }

    /**
     * 返回对象的字符串值，当对象为null时，按照默认值返回
     */
    public static String toString(Object o, String nullDefault) {
        return (o != null) ? o.toString() : nullDefault;
    }

    /**
     * 两参数如果同为null或者指向同一个实例，返回0；否则返回c.compare(a, b)
     *
     * 对于有一个参数为null的情况，是否抛出NullPointerException
     * 取决于传递参数的顺序和比较器的实现
     */
    public static <T> int compare(T a, T b, Comparator<? super T> c) {
        return (a == b) ? 0 :  c.compare(a, b);
    }

    /**
     * 非null检查
     */
    public static <T> T requireNonNull(T obj) {
        if (obj == null)
            throw new NullPointerException();
        return obj;
    }

    /**
     * 非null检查，返回自定义消息的NullPointerException
     * 最佳实践 用于检测表单值
     */
    public static <T> T requireNonNull(T obj, String message) {
        if (obj == null)
            throw new NullPointerException(message);
        return obj;
    }

    /**
     * 检测对象是否为null
     */
    public static boolean isNull(Object obj) {
        return obj == null;
    }

    /**
     * 检测对象是否不是null
     */
    public static boolean nonNull(Object obj) {
        return obj != null;
    }

    /**
     * 非null检查，抛出带有自定义信息的NullPointerException
     *
     * 这个方法将错误信息的生成延迟到确实需要错误信息时，会有一定的性能提升
     */
    public static <T> T requireNonNull(T obj, Supplier<String> messageSupplier) {
        if (obj == null)
            throw new NullPointerException(messageSupplier.get());
        return obj;
    }
}
```

