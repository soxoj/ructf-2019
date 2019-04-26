﻿using SharpGeoAPI.Models;

namespace SharpGeoAPI.Storages
{
    public interface IStorage
    {
        AgentInfo GetAgent(string agentId);
        void AddAgent(AgentInfo agentInfo);
    }
}