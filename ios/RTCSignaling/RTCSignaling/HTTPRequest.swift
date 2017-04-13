//
//  HTTPRequest.swift
//  WebRTC-iOS-Client
//
//  Created by Disi A on 3/30/17.
//  Copyright © 2017 Innovation. All rights reserved.
//

import UIKit

fileprivate var kTimeout:TimeInterval = 10
fileprivate var kAllowSelfSignedCertificate: Bool = false

enum HTTPResponseStatus: Int, CustomStringConvertible {
    // Server Events
    case success = 200,
    successCreated = 201,
    
    // Server Error Events
    badRequest = 400,
    unauthorized = 401,
    forbidden = 403,
    methodNotAllowed = 405,
    preconditionFailed = 412,
    tooManyRequest = 429,
    
    serverInternalError = 500,
    
    // SDK Events
    requestTimeout = 1001,
    sslHandshakeFailure = 1004
    
    var description: String {
        return "\(self)"
    }
}

public class HTTPResponse{
    
    var statusCode:Int
    var responseData:Data
    
    public init(status:Int, response:Data){
        statusCode = status
        responseData = response
    }
    
    public func parseJson() throws -> [String: Any]{
        
        let decoded = try JSONSerialization.jsonObject(with: responseData, options: [])
        // here "decoded" is of type `Any`, decoded from JSON data
        
        // you can now cast it with the right type
        if let json = decoded as? [String: Any] {
            return json
        }else{
            throw NSError(domain: "HTTPResponseError:JsonDecoding", code: 0, userInfo: nil)
        }
    }
    
    public func parseJsonArray() throws -> [Any]{
        let decoded = try JSONSerialization.jsonObject(with: responseData, options: [])
        // here "decoded" is of type `Any`, decoded from JSON data
        
        // you can now cast it with the right type
        if let jsonArray = decoded as? [Any] {
            return jsonArray
        }else{
            throw NSError(domain: "HTTPResponseError:JsonArrayDecoding", code: 1, userInfo: nil)
        }
    }
    
}

public class HTTPRequest {
    
    public static func setTimeout(_ timeout: TimeInterval){
        if timeout > 0{
            kTimeout = timeout
        }else{
            assertionFailure("Setting invalid timeout duration: \(timeout)")
        }
    }
    
    public static func setAllowSelfSignedCertificate(_ allow: Bool){
        kAllowSelfSignedCertificate = allow
    }
    
    public static func post(url:String, json:Dictionary<String, Any>, headers:Dictionary<String, String>, onSuccess:@escaping(_ response: HTTPResponse) -> Void, onFailure:@escaping(_ error:Error) -> Void){
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: json) //, options: .prettyPrinted
            HTTPRequest.sendHTTPRequest(url: url, method: "POST", data: jsonData, headers: headers, onSuccess: onSuccess, onFailure: onFailure, timeout: kTimeout, delegate: nil)
        } catch {
            print(error.localizedDescription)
        }
    }
    
    public static func post(url:String, data:Data, headers:Dictionary<String, String>, onSuccess:@escaping(_ response: HTTPResponse) -> Void, onFailure:@escaping(_ error:Error) -> Void){
        HTTPRequest.sendHTTPRequest(url: url, method: "POST", data: data, headers: headers, onSuccess: onSuccess, onFailure: onFailure, timeout: kTimeout, delegate: nil)
    }
    
    public static func get(url:String, headers:Dictionary<String, String>, onSuccess:@escaping(_ response: HTTPResponse) -> Void,  onFailure:@escaping(_ error:Error) -> Void){
        HTTPRequest.sendHTTPRequest(url: url, method: "POST", data: nil, headers: headers, onSuccess: onSuccess, onFailure: onFailure, timeout: kTimeout, delegate: nil)
    }
    
    fileprivate static func sendHTTPRequest(url:String, method:String, data:Data?, headers:Dictionary<String, String>, onSuccess:@escaping(_ response: HTTPResponse) -> Void, onFailure:@escaping (_ error:Error) -> Void, timeout: TimeInterval, delegate: URLSessionDelegate?){
        let urlAddress = URL(string: url)
        assert(urlAddress != nil, "HTTPRequest: Sending request to invalid URL")
        
        var request = URLRequest(url: urlAddress!, cachePolicy: .reloadIgnoringLocalCacheData, timeoutInterval: timeout)
        request.httpMethod = method
        request.httpBody = data
        
        let sessionConfig = URLSessionConfiguration.default
        sessionConfig.httpAdditionalHeaders = headers
        
        var session:URLSession // = URLSession.shared

        if let sessionDelegate = delegate {
            session = URLSession(configuration: sessionConfig, delegate: sessionDelegate, delegateQueue: nil)
        }else{
            if kAllowSelfSignedCertificate{
                session = URLSession(configuration: sessionConfig, delegate: AllowSelfSignedCertificate(), delegateQueue: nil)
            }else{
                session = URLSession(configuration: sessionConfig, delegate: nil, delegateQueue: OperationQueue.main)
            }
        }
        
        let task = session.dataTask(with: request) { (data, response, error) in
            if let err = error{
                DispatchQueue.main.async {
                    onFailure(err);
                }
            }
            
            var statusCode = HTTPResponseStatus.success.rawValue;

            if let httpResponse = response as? HTTPURLResponse {
                statusCode = httpResponse.statusCode
                
                if statusCode >= 400 {
                    
                    var httpError: Error
                    if let responseStatus = HTTPResponseStatus(rawValue: statusCode){
                        httpError = NSError(domain: responseStatus.description, code: statusCode, userInfo: nil)
                    }else{
                        httpError = NSError(domain: "unknown", code: statusCode, userInfo: nil)
                    }
                    
                    // Optional debug parameter
                    // NSDictionary *jsonData = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
                    // TODO: Refine httpStatusError
                    
                    DispatchQueue.main.async {
                        onFailure(httpError);
                    }
                    return;
                }
            }
            
            // All condition checks are done, call success handler
            
            if let responseData = data{
                DispatchQueue.main.async {
                    onSuccess(HTTPResponse(status: statusCode, response: responseData))
                }
            }else{
                print("WARNING: HTTPRequest: empty response body detected")
                DispatchQueue.main.async {
                    onSuccess(HTTPResponse(status: statusCode, response: Data()))
                }
            }

        }
        task.resume()
        
    }
    
    
}

class AllowSelfSignedCertificate:NSObject, URLSessionDelegate{
    
    func urlSession(_ session: URLSession, didReceive challenge: URLAuthenticationChallenge, completionHandler: @escaping (URLSession.AuthChallengeDisposition, URLCredential?) -> Void){
        completionHandler(.useCredential, URLCredential(trust: challenge.protectionSpace.serverTrust!))
    }

}

public func parseJson(data: Data) throws -> [String: Any]{
    
    let decoded = try JSONSerialization.jsonObject(with: data, options: [])
    // here "decoded" is of type `Any`, decoded from JSON data
    
    // you can now cast it with the right type
    if let json = decoded as? [String: Any] {
        return json
    }else{
        throw NSError(domain: "HTTPRequest:JsonDecoding", code: 0, userInfo: nil)
    }
}
